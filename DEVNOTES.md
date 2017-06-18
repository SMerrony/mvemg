DEVNOTES.md
===========

# MV/Emulator Development Notes #
Things to remember and general thoughts to bear in mind when working on the Go version of the emulator.

## Endianness Issues and Types ##
DG is big-endian, Intel is litle-endian.  Bits are numbered from the left in DG, the right in Intel, so in DG-land
bit 0 is the most significant.

Make use of the dg_word, etc types where possible.

## Emulator Structure ##
### DCH/BMC Map ###
The map overrides memory mapping for certain I/O devices.
We consider it to be 'owned' by the memory module rather than the bus.
Devices which may be subject to DCH/BMC mapping should use the mem...Chan(...) functions to read and write memory.

## Addressing Reminder ##

  * ISZ 32     ; increment the contents of location 32
  * ISZ @32    ; increment the contents of the location addressed by location 32
  * ISZ @32,PC ; increment the contents of the location addressed by (32+PC)
  
## Debugging Strategies ##
  * Lookout for unexpected changes of AC contents around the error
  * Use BREAK command in emulator to stop emulator at a given PC
  * Suspect complicated instructions first!