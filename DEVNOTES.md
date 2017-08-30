DEVNOTES.md
===========

# MV/Emulator Development Notes #
Things to remember and general thoughts to bear in mind when working on the Go version of the emulator.

## Endianness Issues and Types ##
DG is big-endian, Intel is litle-endian.  Bits are numbered from the left in DG, the right in Intel, so in DG-land
bit 0 is the most significant.

Make use of the dg.WordT, etc types where possible.

## Emulator Structure ##

### Memory ###
Memory has been put into its own package primarily to reduce the risk of direct reads and writes to the []ram array.  Minimise what is exported from this package.

### DCH/BMC Map ###
The map overrides memory mapping for certain I/O devices.
We consider it to be 'owned' by the memory module rather than the bus.
Devices which may be subject to DCH/BMC mapping should only use the mem...Chan(...) functions to read and write memory.

### Goroutines (Explicit Threads) ###
  * StatusCollector is mainly a goroutine which waits on status updates and presents them on port 9999
  * Each unit that sends statistics to the StatusCollector has a goroutine dedicated to the task. ie. CPU, DPF, DSKP
  * TTI has a goroutine for the TtiListener which gets console keyboard input
  * DSKP has a goroutine which does the actual disk I/O (CB processing)

## Addressing Reminder ##

  * ISZ 32     ; increment the contents of location 32
  * ISZ @32    ; increment the contents of the location addressed by location 32
  * ISZ @32,PC ; increment the contents of the location addressed by (32+PC)
  
## Debugging Strategies ##
  * Lookout for unexpected changes of AC contents around the error
  * Use BREAK command in emulator to stop emulator at a given PC
  * Suspect complicated instructions first!
  * The DG documentation varies in quality AND content, compare different versions when looking at complex instructions

## Code History/Quality ##
The MTB and SimHTape modules were written very early on (still learning Go) and could do with rewriting at some point.
