# MV/Em User Guide #

The current version of MV/Em emulates a minimally-configured Data General MV/10000 minicomputer from c.1983.

To use the emulator you will need a DASHER compatible terminal emulator with telnet support, 
both [DasherQ](https://github.com/SMerrony/DasherQ) and [DasherJ](https://github.com/SMerrony/DasherJ) are 
known to work well.
The terminal emulator will operate as the master console of the minicomputer and the emulator will not begin 
operation until it is connected.

Depending upon the build of MV/Em, more or less informational and debugging output will be sent to the 
invoking terminal.

## Invocation ##

  `./mvem [doscriptfile]`

When MV/Em is started from a console you may optionally supply a script name which will be executed as an 
initial DO SCRIPT (see below) once a console is attached.

Once you have invoked MV/Em you must connect a DASHER terminal emulator to port 10000 on the local machine.
The emulator will not initialise until there is a connection on that port.

You should immediately be greeted with the welcome message...

  `Welcome to the MV/Emulator - Type HE for help`

If you are running a debug build then a great deal of information about the internal operation of MV/Em will 
appear on the console where you started the emulator.  Redirecting this information to a file instead will 
significantly speed up the operation of the emulator. Eg. 
    `./mvem SCRIPT1.TXT >mvem.log 2>&1`
	
## Emulator Commands ##
MV/Em commands are all entered at the console terminal which behaves rather like the SCP on a real MV/10000 but 
with additional commands to control the emulation; so there are two groups of commands: SCP-CLI commands and Emulator 
commands.

### SCP-CLI Commands ###
These commands are very similar to those provided at a real MV machine (some later additions have been added to 
the original MV/10000 set).

You can break into the SCP when the machine is running by hitting the ESCape key - the machine will pause once 
the current instruction has finished executing.

The following commands have been implemented...

#### . ####
> Display the current state of the CPU, eg. ACs, PC, carry and ATU flags.

#### B `#` ####
> Boot from device number #.  Currently only supports device 22 - the MTB unit.

#### CO ####
> COntinue (or start) processing from the current PC.

#### E A # ####
> Examine/modify Accumulator #.

#### E M # ####
> Examine/modify physical Memory location #.

#### E P ####
> Examine/modify the PC.

#### HE ####
> HElp - display a 1-screen summary of available commands.

#### SS ####
> Single-Step one instruction from the PC.

### Emulator Commands ###
MV/Emulator commands control the emulation environment rather than the virtual machine.  They are loosely based on [[SimH]] commands.

#### ATT `<dev> <file>` ####
> ATTach an image file to the named device.  Tape file images must be in SimH format.  

#### BREAK `<addr>` ####
> Set an execution BREAKpoint at the given address - the emulator will pause if that address is reached.  Use the CO command to continue execution.

#### CHECK ####
> CHECK the validity of an attached tape image by attempting to read it all and displaying a summary of the virtual tape's contents on the console.

#### CREATE `<type> <imageFileName>` ####
> CREATE an empty disk image suitable for attaching to the emulator and initialising with DFMTR.  eg. CREATE DPF BLANK.DPF

#### DIS `<from> <to> | +<#>` ####
> DISplay/disassemble memory between the given addresses or # locations from the PC.

#### DO `<scriptfile>` ####
> DO emulator commands from the file.  
> Here is an example scriptfile which attaches a SimH tape image to the MTB device,  attaches a DPF-type disk image, and
attempts to boot from device 22 (MTB) and finally displays the status of the CPU...

    ATT MTB TAPE1.9trk
	ATT DPF DISK1.DPF
    B 22
    .
  
#### EXIT ####
> EXIT the emulator cleanly.

#### NOBREAK `<addr>`
> Clear any breakpoint at the given address.

#### SHOW DEV ####
> SHOW a display a brief summary all known DEVices and their busy/done flags and statuses
