# TryHackMe SSH Honeypot
SSH Honeypot that gathers attempted creds, IP addresses and versions.
The SSH server will either issue a warning, or drop the attacker into a fake shell.

## Loot File Format
The loot file is a CSV with the following structure:
`Username,Password,Remote Address, Client Version`

## Fake Shell
The fake shell will print a bash command not found error for every command entered, except exit.