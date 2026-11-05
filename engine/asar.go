package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
)

type asarEntry struct {
	path         string
	offset, size int64
	node         map[string]interface{}
}
type asarArchive struct {
	raw        []byte
	tree       map[string]interface{}
	entries    []asarEntry
	start      int64
	headerHash string
}

func hashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func hashFile(p string) (string, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return "", e
	}
	return hashBytes(b), nil
}
func readAsar(p string) (*asarArchive, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return nil, e
	}
	return parseAsar(b)
}
func parseAsar(b []byte) (*asarArchive, error) {
	if len(b) < 16 {
		return nil, errors.New("ASAR too short")
	}
	a, bb, c, d := binary.LittleEndian.Uint32(b[:4]), binary.LittleEndian.Uint32(b[4:8]), binary.LittleEndian.Uint32(b[8:12]), binary.LittleEndian.Uint32(b[12:16])
	if a != 4 || bb != c+4 || c < d+4 || c-d-4 > 3 || uint64(16)+uint64(d) > uint64(len(b)) {
		return nil, errors.New("unsupported ASAR format")
	}
	start := int64(8) + int64(bb)
	if start > int64(len(b)) {
		return nil, errors.New("ASAR data missing")
	}
	tree := map[string]interface{}{}
	if e := json.Unmarshal(b[16:16+d], &tree); e != nil {
		return nil, fmt.Errorf("ASAR JSON: %w", e)
	}
	result := &asarArchive{raw: b, tree: tree, start: start, headerHash: hashBytes(b[16 : 16+d])}
	var visit func(map[string]interface{}, string) error
	visit = func(n map[string]interface{}, prefix string) error {
		children, ok := n["files"].(map[string]interface{})
		if !ok {
			return nil
		}
		for name, value := range children {
			if name == "." || name == ".." || name == "" {
				return errors.New("invalid ASAR entry")
			}
			for _, char := range name {
				if char == '/' || char == '\\' {
					return errors.New("invalid ASAR entry path")
				}
			}
			node, ok := value.(map[string]interface{})
			if !ok {
				return errors.New("bad ASAR entry")
			}
			path := name
			if prefix != "" {
				path = prefix + "/" + name
			}
			if _, ok := node["files"]; ok {
				if e := visit(node, path); e != nil {
					return e
				}
				continue
			}
			if unpacked, ok := node["unpacked"].(bool); ok && unpacked {
				continue
			}
			if _, ok := node["link"]; ok {
				continue
			}
			offS, ok := node["offset"].(string)
			if !ok {
				return fmt.Errorf("ASAR entry %s missing offset", path)
			}
			off, e := strconv.ParseInt(offS, 10, 64)
			if e != nil {
				return e
			}
			sizeF, ok := node["size"].(float64)
			if !ok {
				return fmt.Errorf("ASAR entry %s missing size", path)
			}
			size := int64(sizeF)
			if off < 0 || size < 0 || start+off+size > int64(len(b)) {
				return fmt.Errorf("ASAR entry %s out of range", path)
			}
			result.entries = append(result.entries, asarEntry{path: path, offset: off, size: size, node: node})
		}
		return nil
	}
	if e := visit(tree, ""); e != nil {
		return nil, e
	}
	sort.Slice(result.entries, func(i, j int) bool { return result.entries[i].offset < result.entries[j].offset })
	return result, nil
}
func (a *asarArchive) entry(path string) ([]byte, bool) {
	for _, e := range a.entries {
		if e.path == path {
			return a.raw[a.start+e.offset : a.start+e.offset+e.size], true
		}
	}
	return nil, false
}
func (a *asarArchive) rebuild(changes map[string][]byte) ([]byte, error) {
	// Check all changed entries exist. Keep every other payload byte-for-byte.
	nodes := map[string]bool{}
	for _, e := range a.entries {
		nodes[e.path] = true
	}
	for name := range changes {
		if !nodes[name] {
			return nil, fmt.Errorf("required ASAR entry missing: %s", name)
		}
	}
	var data []byte
	for _, e := range a.entries {
		content := a.raw[a.start+e.offset : a.start+e.offset+e.size]
		if newContent, ok := changes[e.path]; ok {
			content = newContent
		}
		e.node["offset"] = strconv.FormatInt(int64(len(data)), 10)
		e.node["size"] = len(content)
		if i, ok := e.node["integrity"].(map[string]interface{}); ok {
			if alg, ok := i["algorithm"].(string); ok && alg != "SHA256" {
				return nil, fmt.Errorf("unsupported integrity algorithm %s", alg)
			}
			i["hash"] = hashBytes(content)
			if _, ok := i["blocks"]; ok {
				block := int64(4194304)
				if num, ok := i["blockSize"].(float64); ok && num > 0 {
					block = int64(num)
				}
				blocks := []string{}
				for start := int64(0); start < int64(len(content)); start += block {
					end := start + block
					if end > int64(len(content)) {
						end = int64(len(content))
					}
					blocks = append(blocks, hashBytes(content[start:end]))
				}
				i["blocks"] = blocks
			}
		}
		data = append(data, content...)
	}
	hb, e := json.Marshal(a.tree)
	if e != nil {
		return nil, e
	}
	size := len(hb)
	padding := (4 - (size % 4)) % 4
	// Electron ASAR pickle uses: size-pickle (a,b) then string pickle (c,d), padded to four-byte alignment.
	c := uint32(size + 4 + padding)
	bb := c + 4
	prefix := make([]byte, 16)
	binary.LittleEndian.PutUint32(prefix[0:4], 4)
	binary.LittleEndian.PutUint32(prefix[4:8], bb)
	binary.LittleEndian.PutUint32(prefix[8:12], c)
	binary.LittleEndian.PutUint32(prefix[12:16], uint32(size))
	out := make([]byte, 0, 16+size+padding+len(data))
	out = append(out, prefix...)
	out = append(out, hb...)
	for p := 0; p < padding; p++ {
		out = append(out, 0)
	}
	out = append(out, data...)
	verify, e := parseAsar(out)
	if e != nil {
		return nil, e
	}
	for name, b := range changes {
		found, ok := verify.entry(name)
		if !ok || hashBytes(found) != hashBytes(b) {
			return nil, fmt.Errorf("ASAR output mismatch: %s", name)
		}
	}
	return out, nil
}
