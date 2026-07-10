# Gate Display System — Implementation Plan

## Overview

A standalone Electron gate display app for parking entrances, plus backend and management UI.

## Milestones

| # | Milestone | Files | Depends On |
|---|-----------|-------|------------|
| 1 | Backend — Gates Table & Endpoints | 6 | None |
| 2 | Gate Display App — Core Display | 22 | M1 |
| 3 | Gate Display App — Registration Flow | 3 (extends M2) | M2 |
| 4 | Desktop & Dashboard — Gate Management UI | 7 | M1, M3 |

## How to read

- Each milestone file contains: objective, file manifest, implementation details, test plan, manual verification steps
- Gate display app at `gate-display/` (new)
- Backend changes at `backend/`
- Desktop changes at `desktop/`
- Dashboard changes at `dashboard/`

## Architecture

```
LOOP SENSOR ──┐
TICKET BUTTON ─┤
CARD READER ───┤── PERIPHERAL CONTROLLER ── LAN ── GATE DISPLAY APP ── HTTP ── BACKEND API
TICKET DSPNSR ─┤         (PLC / ESP32)                     │                       │
GATE BARRIER  ─┘         (Mock in v1)                       │                       │
                                                            ▼                       ▼
                                                     DESKTOP APP          POSTGRES DB
                                                     (Gate Setup)         (gates table)
                                                                               │
                                                                               ▼
                                                                         DASHBOARD
                                                                         (Gates page)
```
