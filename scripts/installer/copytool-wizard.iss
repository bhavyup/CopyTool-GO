#define AppName "CopyTool"
#define AppPublisher "bhavyup"

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif

#ifndef BinaryPath
  #error BinaryPath define is required
#endif

#ifndef OutputDir
  #define OutputDir "dist"
#endif

[Setup]
AppId={{34CBEBB6-69A6-4DF0-A2DB-8F98A810A8E7}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
DefaultDirName={localappdata}\Programs\copytool
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
OutputDir={#OutputDir}
OutputBaseFilename=copytool-setup_{#AppVersion}_windows_amd64
Compression=lzma
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
ChangesEnvironment=yes
LicenseFile=..\..\LICENSE

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional shortcuts:"; Flags: unchecked

[Files]
Source: "{#BinaryPath}"; DestDir: "{app}\bin"; DestName: "copytool.exe"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\CopyTool"; Filename: "{app}\bin\copytool.exe"
Name: "{autodesktop}\CopyTool"; Filename: "{app}\bin\copytool.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\bin\copytool.exe"; Parameters: "-version"; Flags: postinstall nowait skipifsilent

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{code:GetUpdatedPath|{app}\bin}"; Flags: preservestringtype

[Code]
function NormalizePathEntry(const Value: string): string;
begin
  Result := Lowercase(Trim(Value));
  while (Length(Result) > 0) and ((Result[Length(Result)] = '\') or (Result[Length(Result)] = '/')) do
    Delete(Result, Length(Result), 1);
end;

function PathContainsEntry(const PathValue, Entry: string): Boolean;
var
  SearchPath, SearchEntry: string;
begin
  SearchPath := ';' + Lowercase(PathValue) + ';';
  SearchEntry := ';' + Lowercase(Entry) + ';';
  Result := Pos(SearchEntry, SearchPath) > 0;
end;

function GetCurrentUserPath: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', Result) then
    Result := '';
end;

function GetUpdatedPath(Param: string): string;
var
  CurrentPath: string;
  NormalizedParam: string;
begin
  CurrentPath := GetCurrentUserPath();
  NormalizedParam := NormalizePathEntry(Param);

  if (NormalizedParam = '') then
  begin
    Result := CurrentPath;
    exit;
  end;

  if PathContainsEntry(CurrentPath, NormalizedParam) then
  begin
    Result := CurrentPath;
    exit;
  end;

  if Trim(CurrentPath) = '' then
    Result := Param
  else if CurrentPath[Length(CurrentPath)] = ';' then
    Result := CurrentPath + Param
  else
    Result := CurrentPath + ';' + Param;
end;

function RemovePathEntry(const PathValue, Entry: string): string;
var
  Work, Search: string;
  Position: Integer;
begin
  Work := ';' + PathValue + ';';
  Search := ';' + Lowercase(Entry) + ';';

  Position := Pos(Search, Lowercase(Work));
  while Position > 0 do
  begin
    Delete(Work, Position, Length(Search));
    Position := Pos(Search, Lowercase(Work));
  end;

  while Pos(';;', Work) > 0 do
    StringChangeEx(Work, ';;', ';', True);

  if (Length(Work) > 0) and (Work[1] = ';') then
    Delete(Work, 1, 1);
  if (Length(Work) > 0) and (Work[Length(Work)] = ';') then
    Delete(Work, Length(Work), 1);

  Result := Work;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  CurrentPath, UpdatedPath, BinDir: string;
begin
  if CurUninstallStep <> usUninstall then
    exit;

  BinDir := NormalizePathEntry(ExpandConstant('{app}\bin'));

  if RegQueryStringValue(HKCU, 'Environment', 'Path', CurrentPath) then
  begin
    UpdatedPath := RemovePathEntry(CurrentPath, BinDir);
    RegWriteExpandStringValue(HKCU, 'Environment', 'Path', UpdatedPath);
  end;
end;
