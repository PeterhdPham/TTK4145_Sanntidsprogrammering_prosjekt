# Elevator Software System Checklist

## Main Requirements

### Service Guarantee
- [ ] Ensure hall call button lights turn on, signaling an elevator will arrive.
  - [ ] Trykke på flere samtidit
  - [ ] Ta en og en 
- [ ] Ensure cab call button lights turn on for calls within a specific elevator.
  - [ ] Heisen plukker opp alle sine cab calls

### Reliability
- [ ] Implement handling of failure states without losing calls.
  - [ ] Master skal få tilbake masterlist ved disconnect og connect på nettverk
  - [ ] Master skal få tilbake masterlist ved disconnect og connect på kill 
  - [ ] Client beholder cabcalls ved disconnect og connect på nettverk
  - [ ] Client beholder cabcalls ved disconnect og connect på kill
- [ ] Maintain functionality during network disconnections and recover quickly from failures.
- [ ]

### Functionality
- [ ] Hall call buttons must summon an elevator correctly.
- [ ] Hall button lights should display consistently across all workspaces under normal conditions.
- [ ] Cab and hall button lights turn on promptly after being pressed.
- [ ] Button lights turn off once the corresponding call has been serviced.
- [ ] Implement proper door functionality, including a 3-second open duration and obstruction sensitivity.

### Efficiency
- [ ] Avoid unnecessary stops at every floor.
- [ ] Handle clearing of hall call buttons intelligently, respecting intended travel direction.

## Secondary Requirements
- [ ] Optimize call distribution among elevators for efficiency.

## Permitted Assumptions
- [ ] At least one elevator is operational and can serve calls at all times.
- [ ] Single elevator or disconnected elevators do not require cab call redundancy.
- [ ] No network partitioning scenario will occur.

## Unspecified Behavior (Developer's Discretion)
- [ ] Decide on behavior during initial network connection failures.
- [ ] Define hall button functionality when disconnected from the network.
- [ ] Outline the functionality of the stop button, if implemented.

