protocols-emu
================

A experimental network protocol emulator by GoLang.

With current prototype, the emulator is using a new defined protocol named SSMP to emulate large scale network protocol interaction behavious. Which is simuliar to the access (PPP/DHCP) gateway in real network world.

For the SSMP, please refer to the design doc in design/design.docx.

To use the tool, go to allinone folder,
```bash
cd allinone
go build
./allinone -size <your scale, default is 1000>
```
