param([Parameter(Mandatory=$true)][string]$AppExe,[string]$AppVersion='1.0.0')
$ErrorActionPreference='Stop'
Add-Type -AssemblyName PresentationFramework,PresentationCore,WindowsBase,System.Windows.Forms
[xml]$markup=@'
<Window xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation" xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml"
 Title="LunarPatcher" Width="825" Height="590" MinWidth="760" MinHeight="540"
 WindowStartupLocation="CenterScreen" Background="#FBF9FC" FontFamily="Segoe UI" FontSize="13" Foreground="#282530" UseLayoutRounding="True">
 <Window.Resources>
  <SolidColorBrush x:Key="Accent" Color="#C65F95"/>
  <Style TargetType="Button">
   <Setter Property="Height" Value="34"/><Setter Property="Padding" Value="16,5"/><Setter Property="MinWidth" Value="90"/>
   <Setter Property="Background" Value="#FFFFFF"/><Setter Property="BorderBrush" Value="#DFD6DF"/>
   <Setter Property="Cursor" Value="Hand"/><Setter Property="FontWeight" Value="SemiBold"/>
  </Style>
  <Style x:Key="Primary" TargetType="Button" BasedOn="{StaticResource {x:Type Button}}">
   <Setter Property="Background" Value="#BD568B"/><Setter Property="Foreground" Value="White"/><Setter Property="BorderBrush" Value="#BD568B"/>
  </Style>
  <Style TargetType="RadioButton">
   <Setter Property="Cursor" Value="Hand"/><Setter Property="FontSize" Value="14"/><Setter Property="FontWeight" Value="SemiBold"/>
  </Style>
 </Window.Resources>
 <Grid>
  <Grid.RowDefinitions><RowDefinition Height="64"/><RowDefinition Height="*"/><RowDefinition Height="64"/></Grid.RowDefinitions>
  <Border Grid.Row="0" Background="White" BorderBrush="#ECE5ED" BorderThickness="0,0,0,1">
   <Grid Margin="26,0"><Grid.ColumnDefinitions><ColumnDefinition Width="*"/><ColumnDefinition Width="Auto"/></Grid.ColumnDefinitions>
    <StackPanel Orientation="Horizontal" VerticalAlignment="Center">
     <Border Background="#F5D9E8" CornerRadius="8" Width="34" Height="34" Margin="0,0,12,0"><TextBlock Text="L" FontSize="22" FontWeight="Bold" Foreground="#9F3974" HorizontalAlignment="Center" VerticalAlignment="Center"/></Border>
     <StackPanel VerticalAlignment="Center"><TextBlock Text="LunarPatcher" FontSize="17" FontWeight="SemiBold"/><TextBlock x:Name="VersionLabel" Text="Unofficial launcher customization" FontSize="10.5" Foreground="#817783"/></StackPanel>
    </StackPanel>
    <TextBlock Grid.Column="1" Text="WINDOWS APPLICATION" VerticalAlignment="Center" FontSize="10" FontWeight="SemiBold" Foreground="#A28C9B"/>
   </Grid>
  </Border>
  <Grid Grid.Row="1" Margin="27,23,27,15">
   <Grid.ColumnDefinitions><ColumnDefinition Width="*"/><ColumnDefinition Width="240"/></Grid.ColumnDefinitions>
   <StackPanel Grid.Column="0" Margin="0,0,27,0">
    <TextBlock Text="Customize your launcher" FontSize="23" FontWeight="SemiBold"/>
    <TextBlock Text="Choose an appearance and apply it to your existing Lunar installation." Foreground="#7B727E" FontSize="12" Margin="0,7,0,20" TextWrapping="Wrap"/>
    <Border Background="White" BorderBrush="#EAE2EB" BorderThickness="1" CornerRadius="8" Padding="16,14">
     <StackPanel>
      <RadioButton x:Name="AdsRadio" GroupName="Mode" Content="Remove ads" IsChecked="True"/>
      <TextBlock Text="Original layout without recognized ad slots." Margin="23,5,0,16" FontSize="12" Foreground="#746B76"/>
      <RadioButton x:Name="LayoutRadio" GroupName="Mode" Content="Clean layout"/>
      <TextBlock Text="Remove ads and add compatible spacing improvements." Margin="23,5,0,16" FontSize="12" Foreground="#746B76"/>
      <RadioButton x:Name="SakuraRadio" GroupName="Mode" Content="Sakura"/>
      <TextBlock Text="Fixed pink cherry-blossom background and matching panels." Margin="23,5,0,16" FontSize="12" Foreground="#746B76"/>
      <RadioButton x:Name="ColorRadio" GroupName="Mode" Content="Custom color"/>
      <TextBlock Text="Your own background and matching panels, without Sakura art." Margin="23,5,0,12" FontSize="12" Foreground="#746B76"/>
      <StackPanel Orientation="Horizontal" Margin="22,0,0,0">
       <Border x:Name="ColorPreview" Background="#D676A5" BorderBrush="#C2B6C0" BorderThickness="1" Width="26" Height="26" CornerRadius="4" Margin="0,0,8,0"/>
       <Button x:Name="ChooseColor" Content="Choose color..." Height="29" MinWidth="112" IsEnabled="False"/>
       <TextBlock x:Name="HexLabel" Text="#D676A5" FontSize="11" Foreground="#817683" VerticalAlignment="Center" Margin="10,0,0,0"/>
      </StackPanel>
     </StackPanel>
    </Border>
    <TextBlock Text="Lunar Client installation" FontSize="12" FontWeight="SemiBold" Margin="0,19,0,7"/>
    <Grid><Grid.ColumnDefinitions><ColumnDefinition Width="*"/><ColumnDefinition Width="Auto"/></Grid.ColumnDefinitions>
     <TextBox x:Name="InstallPath" Height="32" Padding="7,4" FontSize="12" VerticalContentAlignment="Center" BorderBrush="#DBD2DE"/>
     <Button x:Name="Browse" Grid.Column="1" Content="Browse" Margin="8,0,0,0" Height="32"/>
    </Grid>
   </StackPanel>
   <StackPanel Grid.Column="1">
    <TextBlock Text="STATUS" FontSize="10" FontWeight="Bold" Foreground="#9E6791" Margin="0,0,0,9"/>
    <Border Background="White" BorderBrush="#EAE2EB" BorderThickness="1" CornerRadius="8" Padding="14">
     <StackPanel>
      <TextBlock x:Name="StatusTitle" Text="Ready" FontWeight="SemiBold" FontSize="15"/>
      <TextBlock x:Name="StatusMessage" Text="Close Lunar Client before applying changes." FontSize="12" Foreground="#786F7B" Margin="0,8,0,0" TextWrapping="Wrap"/>
      <ProgressBar x:Name="Activity" Height="6" Margin="0,14,0,0" IsIndeterminate="True" Foreground="#CE729F" Visibility="Collapsed"/>
     </StackPanel>
    </Border>
    <TextBlock Text="APPLICATION UPDATES" FontSize="10" FontWeight="Bold" Foreground="#9E6791" Margin="0,24,0,9"/>
    <Border Background="White" BorderBrush="#EAE2EB" BorderThickness="1" CornerRadius="8" Padding="14">
     <StackPanel>
      <TextBlock x:Name="UpdateState" Text="Checking GitHub Releases..." TextWrapping="Wrap" FontSize="12" Foreground="#746B76"/>
      <CheckBox x:Name="AutoUpdate" Content="Automatically install updates" IsChecked="True" Margin="0,13,0,12" FontSize="11.5"/>
      <Button x:Name="UpdateButton" Content="Check for updates" Height="30"/>
     </StackPanel>
    </Border>
    <TextBlock Text="Updates affect LunarPatcher, not the official Lunar Client launcher." FontSize="10.5" Foreground="#908593" TextWrapping="Wrap" Margin="0,13,0,0"/>
   </StackPanel>
  </Grid>
  <Border Grid.Row="2" Background="White" BorderBrush="#EAE2EB" BorderThickness="0,1,0,0">
   <Grid Margin="27,12"><Grid.ColumnDefinitions><ColumnDefinition Width="*"/><ColumnDefinition Width="Auto"/><ColumnDefinition Width="Auto"/></Grid.ColumnDefinitions>
    <Button x:Name="Restore" Content="Restore Lunar" HorizontalAlignment="Left" Grid.Column="0"/>
    <Button x:Name="Check" Content="Check compatibility" Grid.Column="1" Margin="0,0,9,0"/>
    <Button x:Name="Apply" Content="Apply selected mode" Grid.Column="2" Style="{StaticResource Primary}" MinWidth="150"/>
   </Grid>
  </Border>
 </Grid>
</Window>
'@
try { $window=[Windows.Markup.XamlReader]::Load((New-Object System.Xml.XmlNodeReader($markup))) }
catch { [System.Windows.MessageBox]::Show($_.Exception.Message,'LunarPatcher UI error')|Out-Null; exit 1 }
$names=@('VersionLabel','AdsRadio','LayoutRadio','SakuraRadio','ColorRadio','ChooseColor','ColorPreview','HexLabel','InstallPath','Browse','Restore','Check','Apply','StatusTitle','StatusMessage','Activity','UpdateState','AutoUpdate','UpdateButton')
foreach($name in $names){Set-Variable -Scope Script -Name $name -Value $window.FindName($name)}
$VersionLabel.Text='Version '+$AppVersion+' · unofficial customizer'
$settingsDir=Join-Path $env:APPDATA 'LunarPatcher'; New-Item -Type Directory -Path $settingsDir -Force | Out-Null
$settingsFile=Join-Path $settingsDir 'settings.json'
$InstallPath.Text=Join-Path $env:LOCALAPPDATA 'Programs\Lunar Client'
$script:color='#D676A5'
$script:job=$null; $script:jobType=''; $script:outFile=''; $script:errFile='';$script:latest=''
if(Test-Path $settingsFile){try{$prefs=Get-Content $settingsFile -Raw|ConvertFrom-Json; if($prefs.path){$InstallPath.Text=$prefs.path};if($prefs.color -match '^#[0-9a-fA-F]{6}$'){$script:color=$prefs.color;$HexLabel.Text=$prefs.color;$ColorPreview.Background=[Windows.Media.BrushConverter]::new().ConvertFromString($prefs.color)}; $AutoUpdate.IsChecked=[bool]$prefs.autoUpdate;switch($prefs.mode){'layout'{$LayoutRadio.IsChecked=$true};'sakura'{$SakuraRadio.IsChecked=$true};'color'{$ColorRadio.IsChecked=$true}}}catch{}}
function Save-Prefs {
 $mode=if($AdsRadio.IsChecked){'ads'}elseif($LayoutRadio.IsChecked){'layout'}elseif($SakuraRadio.IsChecked){'sakura'}else{'color'}
 @{path=$InstallPath.Text;color=$script:color;mode=$mode;autoUpdate=($AutoUpdate.IsChecked -eq $true)} | ConvertTo-Json | Set-Content -Encoding UTF8 $settingsFile
}
function Refresh-Controls {
 $ChooseColor.IsEnabled=($ColorRadio.IsChecked -eq $true) -and !$script:job
 foreach($button in @($Apply,$Restore,$Check,$Browse)){ $button.IsEnabled=(!$script:job) }
 $Activity.Visibility=if($script:job){'Visible'}else{'Collapsed'}
}
foreach($radio in @($AdsRadio,$LayoutRadio,$SakuraRadio,$ColorRadio)){$radio.Add_Checked({ Refresh-Controls })}
$ChooseColor.Add_Click({
 $picker=New-Object System.Windows.Forms.ColorDialog; $picker.FullOpen=$true
 if($picker.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){$script:color=('#{0:X2}{1:X2}{2:X2}' -f $picker.Color.R,$picker.Color.G,$picker.Color.B);$HexLabel.Text=$script:color;$ColorPreview.Background=([Windows.Media.BrushConverter]::new()).ConvertFromString($script:color);Save-Prefs}
})
$Browse.Add_Click({$dialog=New-Object System.Windows.Forms.FolderBrowserDialog;$dialog.Description='Select your Lunar Client installation';if(Test-Path $InstallPath.Text){$dialog.SelectedPath=$InstallPath.Text};if($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){$InstallPath.Text=$dialog.SelectedPath;Save-Prefs}})
function Start-Job([string]$kind,[string]$arguments) {
 if($script:job){return}
 $id=[guid]::NewGuid().ToString('N');$script:outFile=Join-Path $env:TEMP ('lunarpatcher-'+$id+'.out');$script:errFile=Join-Path $env:TEMP ('lunarpatcher-'+$id+'.err')
 $p=New-Object System.Diagnostics.Process
 $p.StartInfo.FileName=$AppExe;$p.StartInfo.Arguments=$arguments;$p.StartInfo.UseShellExecute=$false;$p.StartInfo.CreateNoWindow=$true
 $p.StartInfo.RedirectStandardOutput=$true;$p.StartInfo.RedirectStandardError=$true
 try{[void]$p.Start();$script:job=$p;$script:jobType=$kind;$StatusTitle.Text='Working...';$StatusMessage.Text=if($kind -like 'update*'){'Checking or downloading a GitHub release...'}else{'Processing Lunar installation...'};Refresh-Controls;$timer.Start()}
 catch{ $StatusTitle.Text='Unable to start';$StatusMessage.Text=$_.Exception.Message }
}
$timer=New-Object Windows.Threading.DispatcherTimer;$timer.Interval=[TimeSpan]::FromMilliseconds(400)
$timer.Add_Tick({
 if(!$script:job -or !$script:job.HasExited){return};$timer.Stop()
 $code=$script:job.ExitCode;$out=$script:job.StandardOutput.ReadToEnd();$err=$script:job.StandardError.ReadToEnd()
 $kind=$script:jobType;$script:job.Dispose();$script:job=$null;Refresh-Controls
 if($code -ne 0){$StatusTitle.Text='Operation failed';$StatusMessage.Text=($err+' '+$out).Trim();if($kind -like 'update*'){$UpdateState.Text='Could not check or download updates.'};return}
 if($kind -eq 'update-check'){
  try{$release=$out|ConvertFrom-Json;$script:latest=$release.latest
   if($release.available){$UpdateState.Text='Update available: '+$release.latest; $UpdateButton.Content='Install update';if($AutoUpdate.IsChecked -eq $true){Start-Job 'update-download' '--download-update'}}
   else{$UpdateState.Text='Up to date (v'+$AppVersion+')';$UpdateButton.Content='Check for updates'}
   $StatusTitle.Text='Ready';$StatusMessage.Text='Update check completed.'
  }catch{$UpdateState.Text='Update result was not readable.'}
 }elseif($kind -eq 'update-download'){
  $setupPath=$out.Trim()
  if(!(Test-Path $setupPath)){$UpdateState.Text='The downloaded installer is missing.';return}
  $UpdateState.Text='Starting verified update installer...'
  Save-Prefs
  try{Start-Process -FilePath $setupPath -ArgumentList '/SILENT','/NORESTART','/CLOSEAPPLICATIONS' -ErrorAction Stop; $window.Close()}catch{$StatusTitle.Text='Update launch failed';$StatusMessage.Text=$_.Exception.Message}
 }else{
  $StatusTitle.Text='Completed'
  if($kind -eq 'restore') {$StatusMessage.Text='Original Lunar files restored.'}
  else {
   try {
    $report=$out|ConvertFrom-Json
    $headline=if($kind -eq 'check'){'Compatible. No files changed.'}else{'Patch applied. Open Lunar normally.'}
    $StatusMessage.Text=$headline+' '+$report.announcements
   } catch {$StatusMessage.Text='Completed. '+$out.Trim()}
  }
  Save-Prefs
 }
})
function Invoke-Patch([string]$kind){
 $folder=$InstallPath.Text.Trim()
 if(!(Test-Path (Join-Path $folder 'Lunar Client.exe')) -or !(Test-Path (Join-Path $folder 'resources\app.asar'))){[System.Windows.MessageBox]::Show($window,'Choose the Lunar installation folder containing Lunar Client.exe and resources\app.asar.','Lunar Client not found')|Out-Null;return}
 $mode=if($AdsRadio.IsChecked){'ads'}elseif($LayoutRadio.IsChecked){'layout'}elseif($SakuraRadio.IsChecked){'sakura'}else{'color'}
 $args=if($kind -eq 'restore'){'--restore'}else{'--mode '+$mode+' --color '+$script:color}
 if($kind -eq 'check'){$args+=' --check'}
 $args+=' --dir "'+$folder.Replace('"','')+'"'
 Save-Prefs;Start-Job $kind $args
}
$Apply.Add_Click({Invoke-Patch 'apply'});$Restore.Add_Click({Invoke-Patch 'restore'});$Check.Add_Click({Invoke-Patch 'check'})
$UpdateButton.Add_Click({if($script:latest -and $UpdateButton.Content -eq 'Install update'){Start-Job 'update-download' '--download-update'}else{Start-Job 'update-check' '--check-update'}})
$AutoUpdate.Add_Checked({Save-Prefs});$AutoUpdate.Add_Unchecked({Save-Prefs})
$window.Add_Closing({param($s,$e) if($script:job){$e.Cancel=$true;[System.Windows.MessageBox]::Show($window,'Wait for the current operation to complete.','LunarPatcher')|Out-Null}else{Save-Prefs}})
Refresh-Controls
$window.Add_ContentRendered({Start-Job 'update-check' '--check-update'})
[void]$window.ShowDialog()
