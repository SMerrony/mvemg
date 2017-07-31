BUGS.md
=======

~~# DFMTR DPF Bug #1~~

~~Apparent memory corruption during Pattern 1 of surface analysis.~~

~~Instruction at 11514 STA 0,0,AC3 is overwriting code, causing a crash (JMP to location zero).~~

~~Looking at the code, it might just be that the BBT is overflowing, in which case the disk reading/writing is not working
as expected.~~

# DFMTR DPF Bug #2

When NOT doing a surface analysis  - very similar to DSKP Bug #1

Crash with fatal internal error (not JMP to location zero).

Reports PC: 35502, AC0: 354, AC1: 111021, AC2: 5

As below, several 16-sector blocks appear to be written, then a 25-sector one after which the crash occurs.

# DFMTR DSKP Bug #1

Crash with fatal internal error (not JMP to location zero).

Reports PC: 35502, AC0: 53, AC1: 111412, AC2: 21

The code looks like it could be better structured to distinguish SYNC and ASYNC operations and result statuses.

DSKP Command sequence is:

```
BEGIN
GET MAPPING, SET MAPPING
GET IFACE INFO, SET IFACE INFO
SET CONTROLLER INFO
GET UNIT INFO, SET UNIT INFO
START LIST: RECALIBRATE
START LIST: READ 1 SECT FROM SECT 2
START LIST: WRITE 1 SECT TO SECT 2
GET UNIT INFO
START LIST: RECALIBRATE
START LIST: WRITE 1 SECT TO SECT 3
START LIST: READ 1 SECT FROM SECT 3
START LIST: WRITE 16 SECTS TO SECT 430381
START LIST: WRITE 16 SECTS TO SECT 430397
START LIST: WRITE 16 SECTS TO SECT 430413
START LIST: WRITE 16 SECTS TO SECT 430429
START LIST: WRITE 16 SECTS TO SECT 430445
START LIST: WRITE 16 SECTS TO SECT 430461
START LIST: WRITE 16 SECTS TO SECT 430477
START LIST: WRITE 16 SECTS TO SECT 430493
START LIST: WRITE 16 SECTS TO SECT 430509
START LIST: WRITE 16 SECTS TO SECT 430525
START LIST: WRITE 16 SECTS TO SECT 430541
START LIST: WRITE 16 SECTS TO SECT 430557
START LIST: WRITE 16 SECTS TO SECT 430573
START LIST: WRITE 16 SECTS TO SECT 430589
START LIST: WRITE 16 SECTS TO SECT 430605
START LIST: WRITE 16 SECTS TO SECT 430621
START LIST: WRITE 16 SECTS TO SECT 430637
START LIST: WRITE **25** SECTS TO SECT 430653

```
Each write is followed by a DIC and a DICC, then one more DIC and the DOA/B/CS instructions.

The final write only performs the DIC and DICC.

Exactly the same crash occurs after (apparently succesfully) running a surface analysis.

Looking at the disassembly, there is only space reserved for 16 (or 17) 512B blocks, so why is DFMTR trying to write 25???  Something is going wrong before the final WRITE.

