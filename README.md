# G15Manager: An open source replacement to manage your Asus Zephyrus G15

![Build Release](https://github.com/NeilSeligmann/G15Manager/actions/workflows/release.yml/badge.svg)

## Table of Contets
- [G15Manager: An open source replacement to manage your Asus Zephyrus G15](#g15manager-an-open-source-replacement-to-manage-your-asus-zephyrus-g15)
	- [Table of Contets](#table-of-contets)
	- [Disclaimer](#disclaimer)
	- [Current Status](#current-status)
	- [Web UI](#web-ui)
	- [Bug Report](#bug-report)
	- [Requirements](#requirements)
		- [Technical Notes](#technical-notes)
	- [Install](#install)
	- [Thermal Profiles](#thermal-profiles)
	- [Changing the Fan Curves](#changing-the-fan-curves)
	- [Change Refresh Rate](#change-refresh-rate)
	- [Hotkeys](#hotkeys)
	- [Battery Charge Limit](#battery-charge-limit)
	- [How to Build](#how-to-build)
	- [Developing](#developing)
	- [References](#references)
	- [Credits](#credits)

## Disclaimer

Work in progress. This may void your warranty, proceed at your own risk.

## Current Status

The project is currently under development.
Most of the current features come from the original [G14Manager](https://github.com/zllovesuki/G14Manager)

> The application can be used without a client but for more advanced configurations, you will need to use one.

Current Features:
- Toggle microphone mute/unmute
- Toggle touchpad
- Keyboard brightness adjustment
- [Thermal profile switching](#thermal-profiles)
- [Fan curve control](#changing-the-fan-curves)
- On-screen display
- Web Socket API
- [Web UI](https://github.com/NeilSeligmann/G15Manager-client)
- De-Noising AI (Armoury Crate files are required)

## Web UI

> When the G15 Manager is first launched it will automatically download the latest client from it's [repository](https://github.com/NeilSeligmann/G15Manager-client)

You can open the Web UI by pressing the ROG Key only once (by default), or by going to [http://127.0.0.1:34453/](http://127.0.0.1:34453/).

From there you can easily change any setting you want.


## Bug Report

If you encounter an issue with the G15Manager (e.g. does not start, stuff does not work, etc), please download the debug build `G15Manager.debug.exe`, and run the binary in a Terminal with Administrator Privileges, then submit an issue with the full logs.

## Requirements

- A Zephyrus G15 😏
- Asus Optimization installed (you will need to disable it)

`Asus Optimization` provides the necessary drivers (aka `atkwmiacpi64`). You may check and see if `C:\Windows\System32\DriverStore\FileRepository\asussci2.inf_amd64_xxxxxxxxxxxxxxxx` exists.

G15Manager will most probably not work on other Zephyrus variants. If you have a G14 use [this manager instead](https://github.com/zllovesuki/G15Manager).

Tested Models:
- Zephyrus G15
  - GA503QR

- Zephyrus G14
  - [GA401QM](https://github.com/NeilSeligmann/G15Manager/issues/1) Reported by [@aminoa](https://github.com/aminoa)

Asus Optimization (the service) **cannot** be running, otherwise G15Manager and Asus Optimization will be fighting over control. We only need Asus Optimization (the driver) to be installed so Windows will load `atkwmiacpi64.sys`, and thus expose a `\\.\ATKACPI` device to be used.

You do not need any other software from Asus (e.g. Armoury Crate, MyAsus, etc) running to use G15Manager; you can safely uninstall them from your system. However, some software (e.g. Asus Optimization) are installed as Windows Services, and you should disable them in Services as they would not provide any value:

![Running Services](images/services.png)

>In order to use the De-Noising AI, you must keep the folder ``DenoiseAIPlugin`` from Armoury Crate, then point the G15Manager to the executable ``ArmouryCrate.DenoiseAI.exe`` inside that folder.

### Technical Notes

"ASUS System Control Interface V2" exists as a placeholder so Asus Optimization can have a device "attached" to it, and loads `atkwmiacpi64.sys`. The hardware for ASCI is a stud in the DSDT table.

"Armoury Crate Control Interface" also exists as a placeholder (stud in the DSDT table), and I'm not sure what purpose does this serve. Strictly speaking, you may disable this in Device Manager and suffer no ill side effects.

Only two pieces of hardware are useful for taking full control of your G15: "Microsoft ACPI-Compliant Embedded Controller" (this stores the configuration, including your fan curves), and "Microsoft Windows Management Interface for ACPI" (this interacts with the embedded controller in the firmware). Since they are ACPI functions, user-space applications cannot invoke those methods (unless we run WinRing0). Therefore, `atkwmiacpi64.sys` exists solely to create a kernel mode device (`\\.\ATKACPI`), and an user-space device (`\\DosDevices\ATKACPI`) so user-space applications and interface with the firmware (including controlling the fan curve, among other devious things).

---

Optionally, disable ASUS System Analysis Driver with `sc.exe config "ASUSSAIO" start=disabled` in a Terminal with Administrator privileges, if you do not plan to use MyASUS.

It is recommend to run `G15Manager.exe` on startup using Task Scheduler, don't forget to check "Run with highest privileges".

You can view example Task Scheduler tasks [on this doc](docs/TaskScheduler.md).

## Install

In order to install this app:
- Download the [latest release](https://github.com/NeilSeligmann/G15Manager/releases/latest)
- Drop the desired executable in a folder (Ex. `C:\Programs\G15Manager`)
- Run the executable as an Administrator
- (Optional) Setup Task Scheduler to automatically run the program

After the initial run the G15 Manager will automatically create a folder called "data". This folder will be used to store stuff like the [Web UI](#web-ui) and temporal files.


## Thermal Profiles

When switching Thermal Profiles, the manager will also change the Windows Power Profile.

**Important**: Currently, the default thermal profiles expect Power Plans "High performance" and "Balanced" to be available. If your installation of Windows does not have those Power Plans, make sure to set the correct ones for each thermal profile.


## Changing the Fan Curves

You can change the fan curves for any given profile by using the [Web UI](#web-ui).

Using the `Fn + F5` key combo you can cycle through all the "Fast Switch" profiles. By default: Quiet -> Balanced -> Performance -> Turbo.

## Change Refresh Rate

For battery saving, you can switch the display refresh rate to 60Hz while you are on battery. Use the `Fn + F12` key combo to toggle between 60Hz/165Hz refresh rate on the internal display. You can also do so from the [Web UI](#web-ui).

<!-- ## Automatic Thermal Profile Switching

For the initial release, it is hardcoded to be:

- On power adapter plugged in: "Performance" Profile (with "High Performance" Power Plan)
- On power adapter unplugged: "Balanced" Profile (With "Balanced" Power Plan)

There is a 5 seconds delay before changing the profile upon power source changes.

To enable this feature, pass `-autoThermal` flag to enable it:

```
.\G15Manager.exe -autoThermal
``` -->

## Hotkeys
|      Hotkey      |      Command      |
| ---------------- | ----------------- |
| `ROG Key`  |  Opens the Web UI       |
| `Fn` + `F1`  |  Mute/Unmute Audio      |
| `Fn` + `F2`  | Keyboard Brightness Down|
| `Fn` + `F3`  |  Keyboard Brightness Up |
| `Fn` + `F4`  |  Play/Pause Media       |
| `Fn` + `F5`  |  Cycle Thermal Profiles |
| `Fn` + `F6`  |  Screenshot             |
| `Fn` + `F7`  |  Display Brightness Down|
| `Fn` + `F8`  |  Display Brightness Up  |
| `Fn` + `F9`  |  Display Mirror Settings|
| `Fn` + `F10` | Enable/Disable Touchpad |
| `Fn` + `F11` |  Sleep                  |
| `Fn` + `F12` |  Toggle Refresh Rate    |
| `Fn` + `C`   |  Disable Dedicated GPU  |
| `Fn` + `V`   |  Enable Dedicated GPU   |

## Battery Charge Limit

By default, G15Manager will set the battery limit charge to 60%.

This can be changed using the [Web UI](#web-ui).

## How to Build

1. Install golang 1.14+ if you don't have it already
2. Install mingw x86_64 for `gcc.exe`
2. Install `rsrc`: `go get github.com/akavel/rsrc`
3. Generate `syso` file: `\path\to\rsrc.exe -arch amd64 -manifest G15Manager.exe.manifest -ico go.ico -o G15Manager.exe.syso`
4. Build the binary: `.\scripts\build.ps1`

## Developing

Use `.\scripts\run.ps1`.

Most keycodes can be found in [reverse_eng/codes.txt](https://github.com/zllovesuki/reverse_engineering/blob/master/G14/codes.txt), and the repo contains USB and API calls captures for reference.

## References

- https://github.com/torvalds/linux/blob/master/drivers/platform/x86/asus-wmi.c
- https://github.com/torvalds/linux/blob/master/drivers/platform/x86/asus-nb-wmi.c
- https://github.com/torvalds/linux/blob/master/drivers/hid/hid-asus.c
- https://github.com/flukejones/rog-core/blob/master/kernel-patch/0001-HID-asus-add-support-for-ASUS-N-Key-keyboard-v5.8.patch
- https://github.com/rufferson/ashs
- https://code.woboq.org/linux/linux/include/linux/platform_data/x86/asus-wmi.h.html
- http://gauss.ececs.uc.edu/Courses/c4029/pdf/ACPI_6.0.pdf
- https://wiki.ubuntu.com/Kernel/Reference/WMI
- [DSDT Table](https://github.com/zllovesuki/reverse_engineering/blob/master/G14/g14-dsdt.dsl)
- [Reverse Engineering](https://github.com/zllovesuki/reverse_engineering/tree/master/G14)

## Credits
[zllovesuki](https://github.com/zllovesuki) for the original G14 Manager.

"Go" logo licensed under unsplash license: [https://blog.golang.org/go-brand](https://blog.golang.org/go-brand)

"Dead computer" logo licensed under Creative Commons: [https://thenounproject.com/term/dead-computer/98571/](https://thenounproject.com/term/dead-computer/98571/)
