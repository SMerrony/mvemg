BUGS.md
=======

# DFMTR DPF Bug

Apparent memory corruption during Pattern 1 of disk test.

Instruction at 11514 STA 0,0,AC3 is overwriting code, causing a crash.

Looking at the code, it might just be that the BBT is overflowing, in which case the disk reading/writing is not working
as expected.


# DFMTR DSKP Bug

DFMTR does not like the status/flag responses from the emulated DSKP unit. Some retries are performed then 
it gives up.

The code looks like it could be better structured to distinguish SYNC and ASYNC operations and result statuses.

