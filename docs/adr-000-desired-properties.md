# Desired Properties of Rolling Shutter

- Author: Ralf Schmitt
- Status: proposed
- Date: 2021-07-14

This document lists the desired properties of a shutter implementation based on rollups. At the moment we're trying to come up with a architecture for the final system. We can use this document to decide if our proposed architecture will have those desired properties. In fact we extracted these properties from one of the proposals.

## Required Properties

Rolling Shutter must have the following properties:

- Transactions are frontrunning protected by default, both from unprivileged (users) and privileged (sequencer, collator) participants.
- Smart contracts on the rollup are composable, i.e., everything can call everything (as opposed to on-chain Shutter where interaction with contracts not using Shutter is limited).

## Desired Properties

Rolling Shutter should have the following properties:

- The maximum throughput is comparable to other rollups.
- The time between two blocks is short (~15s).
- The system works with large keyper sets of ~200 members.
- Users can submit unencrypted transactions, thereby still making progress even if decryption keys or sequencer are unavailable. This can have a relatively long time to finality (~hours).
- The system is largely based on an existing open-source rollup implementation so that we don't have to reinvent the wheel.
- Switching to a different rollup implementation is possible with relatively little effort.
- DoS attacks exploiting the fact that transactions are encrypted are futile.
- The user experience is pleasant, in particular with regards to paying transaction fees.
