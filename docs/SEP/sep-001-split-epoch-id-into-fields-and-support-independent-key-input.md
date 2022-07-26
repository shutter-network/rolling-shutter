# Shutter Enhancement Proposal: Split epoch id into fields and support independent key input 

## Current state

Currently, the epoch id is a single field of size uint64 consisting of:
- 4 bytes (uint32): activation block number
- 4 bytes (uint32): epoch sequence number

This value is used both as an identifier as well as the input to generate the epoch keys.

## Proposal

- Split activation block and sequence id into separate fields.
- Add a key source field (name TBD) 

The collator needs to ensure that the key source field never accepts duplicate values. 

## Rationale

During implementation of the Snapshot integration it became apparent that we need to support much larger ids (256 bit).  
The current scheme of merging block and sequence number into a single field make it cumbersome to work with and also will not easily scale to large id values.

The current epoch id also is used as the key input as is. 
This will not work with the snapshot use case since they don't know about activation block numbers and only use their id as the key input.
Therefore it is necessary to provide a separate field that will be used as the key input.
In traditional rolling shutter it will simply contain the same value as the current epoch id. 
For snapshot it will only consist of the provided external id.


