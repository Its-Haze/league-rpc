Unicode true

####
## Installer for League RPC. Values not defined here come from "wails_tools.nsh",
## which is generated from build/config.yml by `task assets:update`.
##
## To build this by hand, outside of Task:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\..\bin\league-rpc.exe project.nsi
####

## Add/Remove Programs key. Without this, wails_tools.nsh concatenates the
## company and product names into "HazeLeague RPC".
!define UNINST_KEY_NAME "LeagueRPC"

## The app's single-instance mutex, from singleInstanceID in cmd/league-rpc-gui.
!define APP_MUTEX "wails-app-com.its-haze.league-rpc-sim"

## Autorun value written by internal/startup; startup.ValueName must match.
!define AUTORUN_KEY   "Software\Microsoft\Windows\CurrentVersion\Run"
!define AUTORUN_VALUE "LeagueRPC"

!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

# Overrides MUI's default welcome text, which warns about closing other apps
# and rebooting for shared system files. This install is per-user, touches no
# shared files, and never needs a reboot, so that warning is just wrong here.
!define MUI_WELCOMEPAGE_TEXT "Setup will guide you through the installation of ${INFO_PRODUCTNAME}.$\r$\n$\r$\nClick Next to continue."

# A machine-scope installer runs elevated, so launching from it would hand the
# app an admin token it should never have.
!if "${WAILS_INSTALL_SCOPE}" == "user"
    !define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
    !define MUI_FINISHPAGE_RUN_TEXT "Start ${INFO_PRODUCTNAME}"
!endif

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
!insertmacro MUI_PAGE_COMPONENTS # Optional components.
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES # Uninstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${INFO_PRODUCTVERSION}-setup.exe" # Name of the installer's file.
!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif
ShowInstDetails show # This will always show the installation details.

# The app holds an open handle on its own .exe, so overwriting it while it runs
# leaves a half-updated install. There is no polite remote-quit, hence the ask.
!macro AbortIfRunning UN
Function ${UN}AbortIfRunning
    retry:
        System::Call 'kernel32::OpenMutex(i 0x00100000, b 0, t "${APP_MUTEX}") i .r0'
        IntCmp $0 0 notRunning
        System::Call 'kernel32::CloseHandle(i $0)'
        MessageBox MB_RETRYCANCEL|MB_ICONEXCLAMATION \
            "${INFO_PRODUCTNAME} is still running.$\n$\nQuit it from the system tray, then click Retry." \
            IDRETRY retry
        Abort
    notRunning:
FunctionEnd
!macroend
!insertmacro AbortIfRunning ""
!insertmacro AbortIfRunning "un."

Function .onInit
   !insertmacro wails.checkArchitecture
   Call AbortIfRunning
FunctionEnd

Function un.onInit
   Call un.AbortIfRunning
FunctionEnd

Section "${INFO_PRODUCTNAME}" SecCore
    SectionIn RO

    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    
    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols
    
    !insertmacro wails.writeUninstaller
SectionEnd

Section /o "Desktop shortcut" SecDesktop
    !insertmacro wails.setShellContext
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
SectionEnd

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
    !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} "${INFO_PRODUCTNAME} and its Start Menu shortcut."
    !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} "Add a shortcut to the desktop."
!insertmacro MUI_FUNCTION_DESCRIPTION_END

Section "uninstall" 
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    # Otherwise Windows keeps launching a deleted .exe at every sign-in.
    DeleteRegValue HKCU "${AUTORUN_KEY}" "${AUTORUN_VALUE}"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
