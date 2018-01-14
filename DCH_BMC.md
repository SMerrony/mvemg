# Data CHannel and Burst Multiplexor Channel
_This document is paraphrased from the AOS/VS Internals Reference Manual for AOS/VS 5.0_

The data channel (DCH) provides I/O communications for medium-speed devices (eg. tape drives) and synchronous communication.  The burst multiplexor channel (BMC) is a high speed communications pathwat that transfers data directly between main memory and high-speed peripherals (eg. disk drives).  **The I/O-to-memory transfers for both DCH and BMC always bypass the address translator.**

## DCH/BMC Maps
A map controls a DCH or BMC.  This map is a series of contiguous map slots, each of which contains a pair of map registers - an even-numbered register and its corresponding odd-numbered register.

An MV computer supports 16 DCH maps, each of which contains 32 map slots.  The DCH sends to the processor a logical address with each data transfer.  The processor translates the logical address into a physical address using the appropriate map slot for that address.

The device controller performing the data transfer controls the BMC.  No program control or CPU interaction is required, except when setting up the BMC's map table.  The BMC has two address modes and contains its own map.

### BMC Address Modes
The BMC operates in either the unmapped mode - i.e. the physical mode, or the mapped mode - i.e. the logical mode.

In the unmapped mode, the BMC receives 20-bit addresses from the device controllers and passes them directly to memory.  As the BMC transfers each data word to or from memory, it increments the destination address, causing successive words to move to or from consecutive locations in memory.

If the controller specifies the mapped mode for data transfer, the high-order 10 bits of the logical address form a logical page number, which the BMC map translates into a 10-bit physical page number.  This page number, combined with the 10 low-order bits from the logical address, forms a 20-bit physical address, which the BMV uses to access memory.

## BMC Map
The BMC uses its own map to translate logical page numbers into physical ones.  The map contains 1024 map registers, the odd-numbered registers each containing a 10-bit physical page number.  The BMC uses the logical page number as an index into the map table, and the contents of the selected map register becomes the high-order 10 bits of the physical address.

Note that when the BMC performs a mapped transfer, it increments the destination address after it moves each data word.  If the increment causes an overflow out of the 10 low-order bits, this selects a new map register for subsequent address translation.  Depending upon the contents of the map table, this could mean that the BMC cannot transfer successive words to or from consecutive pages in memory.

## DCH/BMC Registers
An MV computer system contains 512 DCH registers and 1024 BMC registers.  The map registers are numbered from 0 through 07777.

Registers | Description
----------|------------
0000 - 3776 | Even-numbered regs, most significant half of BMC map posns 0 - 1777
0001 - 3777 | Odd-numbered regs, least significant half or BMC map posns 0 - 1777
4000 - 5776 | Even-numbered regs, most significant half of DCH map posns 0 - 777
4001 - 5777 | Odd-numbered regs, least significant half of DCH maps posns 0 - 777
6000        | I/O channel definition register
6001 - 7677 | (reserved)
7700        | I/O channel status register
7701        | I/O channel mask register
7702 - 7777 | (reserved)

### Even-Numbered Register Format
V | D | Hardware Reserved
--|---|------------------
0 | 1 |2 -             15

V - validity bit; if 1 then processor denies access
D - data bit; if 0 the channel transfers data, if 1 the channel transfers zeroes
Reserved should be written to with zeroes; reading these returns an undefined state.

### Odd-Numbered Register Format
Res | Physical Page Number
----|---------------------
0-1 | 2 - 15

Res - hardware reserved
Physical Page Number - associated with logical page reference.

### I/O Channel Definition Register Format
E | Res | BV | DV | Res | BX | A | P | Dis | I/O Channel | M | 0
--|-----|----|----|-----|----|---|---|-----|-------------|---|--
0 | 1-2 | 3  | 4  | 5   | 6  | 7 | 8 | 9   | 10-13       |14 | 15

  * E - error flag
  * Res - reserved
  * BV - BMC validity error flag, if 1 BMC protect error has occurred
  * DV - DCH validity error flag, if 1 DCH protect errro has occurred
  * BX - BMC transfer flag, if 1 and BMC transfer is in progress
  * A - BMC address parity error
  * P - BMC data parity error
  * DIS - disable block transfer
  * I/O Channel - I/O channel number
  * M - DCH mode, if 1 DCH mapping is enabled
  * 0 - always set to 0

### I/O Channel Status Register Format
E | Res | XDCH | 1 | MSK | INT
--|-----|------|---|-----|----
0 | 1-11| 12   |13 | 14  | 15

  * E - error flag
  * Res - reserved
  * XDCH - DCH map slots and operations supported
  * 1 - always set to 1
  * MSK - prevents all devices connected to channel from interrupting the CPU
  * INT - Interrupt pending

### I/O Channel Mask Register Format (MV/10000)
Res | MK0 | MK1 | Res
----|-----|-----|----
0-7 | 8   | 9   | 10-15

  * Res - reserved
  * MK0 - prevents all devices on channel 0 from interrupting CPU
  * MK1 - prevents all devices on channel 1 from interrupting CPU
