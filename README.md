# TryHackMe SSH Honeypot
SSH Honeypot that gathers attempted creds, IP addresses and versions.
The SSH server will either issue a warning, or drop the attacker into a fake shell.

## Loot File Format
The logging now logs to an SQLite database, with schema available in database.sql

## Fake Shell
The fake shell will print a bash command not found error for every command entered, except exit.
You can enable logging of these commands with the -C flag.
