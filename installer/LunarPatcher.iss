#define MyAppName "LunarPatcher"
#define MyAppVersion GetEnv("APP_VERSION")
#if MyAppVersion == ""
  #define MyAppVersion "1.0.0"
#endif
[Setup]
AppId={{B43FBCA2-459F-41D2-B724-A15637816E55}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher=ashclient
AppPublisherURL=https://github.com/ashclient/lunarpatcher
DefaultDirName={localappdata}\Programs\LunarPatcher
DefaultGroupName=LunarPatcher
UninstallDisplayIcon={app}\LunarPatcher.exe
OutputDir=dist
OutputBaseFilename=LunarPatcher_Setup
Compression=lzma
SolidCompression=yes
PrivilegesRequired=lowest
WizardStyle=modern
DisableProgramGroupPage=yes
CloseApplications=yes
RestartApplications=no
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
SetupLogging=yes
[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional shortcuts:"; Flags: unchecked
[Files]
Source: "dist\LunarPatcher.exe"; DestDir: "{app}"; Flags: ignoreversion
[Icons]
Name: "{group}\LunarPatcher"; Filename: "{app}\LunarPatcher.exe"
Name: "{autodesktop}\LunarPatcher"; Filename: "{app}\LunarPatcher.exe"; Tasks: desktopicon
[Run]
Filename: "{app}\LunarPatcher.exe"; Description: "Launch LunarPatcher"; Flags: nowait postinstall skipifsilent
