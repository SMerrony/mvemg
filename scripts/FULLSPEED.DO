# FULLSPEED.DO
# Comments begin with the hash symbol in the first column
#
# This script has logging and debugging turned off, so the emulator runs
# at full speed but does not emit much info when errors 
# or unimplemented features occur.
#
# Logging slows down the eumlator by approx. 9x
SET LOGGING OFF
#
# Attach a SimH-standard tape image
ATT MTB test.9trk
#
# Attach a 6061-type disk image
ATT DPF DISK1.DPF
#
# and a 6239-type disk image
ATT DSKP DISK1.DSKP
#
# Boot device 22 (MTB) - this doesn't start the CPU yet
B 22
# and COntinue to commence the run...
CO