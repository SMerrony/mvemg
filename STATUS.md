# MVEmg Status

* Last advance: 3rd Oct 2018
* Last updated: 2nd Nov 2019

## What Works?

(From home-made 7.73 system tape image)

* File 0 - TBOOT - loads and runs OK
* File 2 - DFMTR - seems to format and verify both type 6061 and type 6239 disks
* File 3 - INSTL - appears to run to completion on 6061 DPF type disk (29th April 2018) and 6239 DPJ type disk (3rd Oct 2018)

## What's Next?

* Boot starter system from disk (!)

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

* File 1 - FIXUP - hangs during startup

  Crash at 347753 decoded as XJMP @2,PC 

  0347753: CE 09 147011 11001110 00001001 "  " XJMP @2.,PC [2-Word OpCode] 
  0347754: 80 01 100001 10000000 00000001 "  "                                    
  0347755: 00 01 000001 00000000 00000001 "  " JMP  1.                            
  0347756: D0 40 150100 11010000 01000000 " @" COM L  2,2                         
  0347757: 00 00 000000 00000000 00000000 "  " JMP  0.                            
  0347760: 00 00 000000 00000000 00000000 "  " JMP  0. 

  ...reading beyond physical memory  
