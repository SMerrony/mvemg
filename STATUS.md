# MVEmg Status

* Last updated: 28th Dec 2019
* Last significant progress: 3rd Oct 2018

## What Works?

(From home-made 7.73 system tape image)

* File 0 - TBOOT - loads and runs OK
* File 2 - DFMTR - seems to format and verify both type 6061 and type 6239 disks
* File 3 - INSTL - appears to run to completion on 6061 DPF type disk (29th April 2018) and 6239 DPJ type disk (3rd Oct 2018)

## What's Next?

I currently have three standalone binaries: SYSBOOT(!), FIXUP, and PCOPY...

### Boot starter system from disk (!)

  Does a countdown then jumps to empty location zero from 77010 via 1001

  077002: 2A 00 025000 00101010 00000000 "* " LDA 1, 0.,AC2                       
  077003: 4B 00 045400 01001011 00000000 "K " STA 1, 0.,AC3                       
  077004: D3 00 151400 11010011 00000000 "  " INC    2,2                          
  077005: FB 00 175400 11111011 00000000 "  " INC    3,3                          
  077006: 19 03 014403 00011001 00000011 "  " DSZ  3.,PC                          
  077007: 01 FB 000773 00000001 11111011 "  " JMP  -5.,PC                         
  077010: 04 02 002002 00000100 00000010 "  " JMP @2.                             
  077011: 00 00 000000 00000000 00000000 "  " JMP  0.                             
  077012: 02 01 001001 00000010 00000001 "  " JMP  1.,AC2  

  01000: 00 00 000000 00000000 00000000 "  " JMP  0.                              
  01001: 00 00 000000 00000000 00000000 "  " JMP  0.                              
  01002: 00 00 000000 00000000 00000000 "  " JMP  0.

  Page 0 looks to have been corrupted as it contains text fragments.

### Standalone FIXUP - Tape file 1 - crash during startup

  Crash at 330020...reading beyond physical memory in XNLDA 0, @-11, AC3
  AC3 contains 30556
  30556 - 11 = 30545
  Location 30545 contains 1432, location 30546 contains 100000
 
### Standalone PCOPY - Tape file 1
  "Fatal disk error, device 27 00, Status = 000000 000000
   Abort!"
   ...repeated forever.  Interestingly, no activity is recorded in dpf_debug.log.
   PC is around 025462

   Also, there are some 'funnies' in the PCOPY start-up prompts.  The REV number is displayed as "00070073◊◊◊.◊◊◊", "Specify source LDU" is missing the first character, "Disk unit name: " is also missing the first character.  Running a 'strings' command on the PCOPY binary shows that these strings are intact in the binary.  _?Could be a byte addressing issue? Check WCMV, WBLM,XLDB, WCLM, WSTB, WLDB._
