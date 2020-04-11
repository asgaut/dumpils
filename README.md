# dumpils

dumpils demodulates ILS signal and dumps the measurements
to stdout in CSV format.

Requires rtl_tcp to be running and listening on localhost:1234.

## Install RTL-SDR driver (Windows)

https://osmocom.org/projects/rtl-sdr/wiki

Download and install under localappdata\programs\rtl-sdr can be done as follows:

```powershell
$name = "rtl-sdr-64bit-20200405"
$dest = "$Env:LOCALAPPDATA\Programs\rtl-sdr"
New-Item -ItemType "directory" -Path $dest -Force
Invoke-WebRequest "https://ftp.osmocom.org/binaries/windows/rtl-sdr/$name.zip" -outfile "$name.zip"
Expand-Archive "$name.zip" -DestinationPath $dest
Move-Item "$dest\$name\*" -Destination $dest
Remove-Item "$dest\$name", "$name.zip"
[Environment]::SetEnvironmentVariable("Path", $Env:Path+";"+$dest, "User")
```

Run zadig to disable the built in driver for RTL-SDR: https://zadig.akeo.ie/

## Running

dumpils reads measurements from a rtl_tcp server. Start rtl_tcp.exe first.

The ILS channel is by default 110.1 MHz.

Run with ```go run cmd/dumpils/main.go```

## Example
```text
C:\> .\dumpils.exe
RF(dbFS);DDM(uA);SDM(%);Ident
-4.4;-0.069;40.125;0.009
-4.3;111.194;18.989;0.193
-4.4;0.105;40.165;0.004
-4.4;-0.080;40.172;0.005
-4.4;0.034;40.171;0.006
-4.4;-0.032;40.174;0.000
-4.4;0.057;40.167;0.005
-4.4;-0.009;40.170;0.007
-4.4;0.071;40.166;0.003
-4.4;-0.038;40.181;0.002
-4.4;-0.047;40.174;0.006
-4.4;-0.022;40.178;0.002
-4.4;-0.012;40.168;0.001
-4.4;0.004;40.170;0.004
-4.4;0.039;40.185;0.006
-4.4;-0.034;40.178;0.003
-4.4;-0.015;40.177;0.006
-4.4;-0.029;40.174;0.004
-4.4;-0.035;40.172;0.003
Exiting on Ctrl-C.
```