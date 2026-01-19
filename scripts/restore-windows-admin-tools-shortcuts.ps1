<#
.SYNOPSIS
    Restores Windows administrative and system tools shortcuts to Start Menu folders.

.DESCRIPTION
    This PowerShell script creates and maintains organized shortcuts for Windows system administration
    and diagnostic tools in the Start Menu. It organizes shortcuts into three categories:
    
    - Windows Tools: Common system utilities (Computer Management, Event Viewer, Services, etc.)
    - Administrative Tools: Advanced system configuration tools (Group Policy, Security Policy, etc.)
    - System Tools: General system utilities (Control Panel, Task Manager, Snipping Tool, etc.)
    
    The script will:
    1. Create the necessary folder structure if it doesn't exist
    2. Check if shortcuts already exist (won't overwrite existing ones)
    3. Create shortcuts pointing to the appropriate system commands/executables
    4. Set proper working directories for each shortcut
    
    This is useful after clean Windows installations or when shortcuts have been deleted, allowing
    easy access to administrative tools without needing to search or navigate manually.

.NOTES
    - Requires administrator privileges to create shortcuts in ProgramData directory
    - Shortcuts are created only if they don't already exist
    - Working directory is set to C:\Windows\System32 for all shortcuts
#>

# === Paths ===
$programs = "C:\ProgramData\Microsoft\Windows\Start Menu\Programs"
$windowsTools = Join-Path $programs "Windows Tools"
$adminTools = Join-Path $programs "Administrative Tools"
$systemTools = Join-Path $programs "System Tools"

# Ensure folders exist
foreach ($folder in @($windowsTools, $adminTools, $systemTools)) {
    if (!(Test-Path $folder)) { New-Item -ItemType Directory -Path $folder | Out-Null }
}

# === Shortcut definitions ===
$shortcuts = @{

    # Windows Tools
    "$windowsTools\Computer Management.lnk"                   = "compmgmt.msc"
    "$windowsTools\Disk Cleanup.lnk"                          = "cleanmgr.exe"
    "$windowsTools\Event Viewer.lnk"                          = "eventvwr.msc"
    "$windowsTools\Services.lnk"                              = "services.msc"
    "$windowsTools\System Information.lnk"                    = "msinfo32.exe"
    "$windowsTools\Task Scheduler.lnk"                        = "taskschd.msc"
    "$windowsTools\Performance Monitor.lnk"                   = "perfmon.exe"
    "$windowsTools\Resource Monitor.lnk"                      = "resmon.exe"
    "$windowsTools\Defragment and Optimize Drives.lnk"        = "dfrgui.exe"
    "$windowsTools\Windows Memory Diagnostic.lnk"             = "mdsched.exe"
    "$windowsTools\Registry Editor.lnk"                       = "regedit.exe"
    "$windowsTools\Command Prompt.lnk"                        = "cmd.exe"
    "$windowsTools\Windows PowerShell.lnk"                    = "powershell.exe"
    "$windowsTools\Windows Terminal.lnk"                      = "wt.exe"

    # Administrative Tools
    "$adminTools\Local Security Policy.lnk"                   = "secpol.msc"
    "$adminTools\Local Group Policy Editor.lnk"               = "gpedit.msc"
    "$adminTools\ODBC Data Sources (32-bit).lnk"              = "odbcad32.exe"
    "$adminTools\ODBC Data Sources (64-bit).lnk"              = "C:\Windows\System32\odbcad32.exe"
    "$adminTools\Print Management.lnk"                        = "printmanagement.msc"
    "$adminTools\Windows Firewall with Advanced Security.lnk" = "wf.msc"
    "$adminTools\iSCSI Initiator.lnk"                         = "iscsicpl.exe"

    # System Tools
    "$systemTools\Control Panel.lnk"                          = "control.exe"
    "$systemTools\System Configuration (msconfig).lnk"        = "msconfig.exe"
    "$systemTools\Character Map.lnk"                          = "charmap.exe"
    "$systemTools\Snipping Tool.lnk"                          = "snippingtool.exe"
    "$systemTools\Task Manager.lnk"                           = "taskmgr.exe"
}

# === Create shortcuts ===
$WScriptShell = New-Object -ComObject WScript.Shell

foreach ($lnk in $shortcuts.Keys) {
    if (!(Test-Path $lnk)) {
        $target = $shortcuts[$lnk]
        $shortcut = $WScriptShell.CreateShortcut($lnk)
        $shortcut.TargetPath = $target
        $shortcut.WorkingDirectory = "C:\Windows\System32"
        $shortcut.Save()
    }
}

Write-Host "All Windows Tools, Administrative Tools, and System Tools shortcuts restored."