# MVEmg Status

Updated 3rd October 2018

## What Works?

(From home-made 7.73 system tape image)

* File 0 - TBOOT - loads and runs OK
* File 2 - DFMTR - seems to format and verify both type 6061 and type 6239 disks
* File 3 - INSTL - appears to run to completion on 6061 DPF type disk (29th April 2018) and 6239 DPJ type disk (3rd Oct 2018)

## What's Next?

* Boot starter system from disk (!)

* File 1 - FIXUP - hangs during startup

  Jumps to empty location zero from 347753 decoded as XJMP @2,PC 

  0347753: CE 09 147011 11001110 00001001 "  " XJMP @2.,PC [2-Word OpCode] 
  0347754: 80 01 100001 10000000 00000001 "  "                                    
  0347755: 00 01 000001 00000000 00000001 "  " JMP  1.                            
  0347756: D0 40 150100 11010000 01000000 " @" COM L  2,2                         
  0347757: 00 00 000000 00000000 00000000 "  " JMP  0.                            
  0347760: 00 00 000000 00000000 00000000 "  " JMP  0. 
