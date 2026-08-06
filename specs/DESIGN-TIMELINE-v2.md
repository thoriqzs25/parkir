# PARKIR v2 — Design Timeline
**Created:** 2026-08-06

---

## Design Workstreams

### 1. Dashboard Design (AMB Admin Web)
### 2. Gate App Design (Driver-Facing Monitor)
### 3. Server Room App Design (Staff Dashboard)
### 4. Hardware Interaction Design (Physical UI)

---

## Week -2 to -1: Research & Discovery (Before Implementation)

**Duration:** 2 weeks
**Deliverables:**

### User Research
- [ ] Interview AMB staff (2-3 locations)
  - Current workflow (manual or automated?)
  - Pain points with current system
  - Offline SOP (what do they do now?)
  - Staff roles & responsibilities
- [ ] Observe driver behavior (entry/exit flow)
  - How long does entry take?
  - Common issues (QR unreadable, payment failure)?
  - Driver demographics (tech-savvy?)

### Competitive Analysis
- [ ] Analyze AMB's current system (screenshots, flows)
- [ ] Analyze 2-3 competitor parking systems
- [ ] Identify best practices + gaps

### Technical Constraints
- [ ] Hardware specs (monitor resolution, printer capabilities)
- [ ] Network constraints (LAN speed, reliability)
- [ ] Accessibility requirements (driver-facing monitor height, font size)

**Stakeholders:** Product manager, UX researcher, AMB operations team

---

## Week 1-2: Information Architecture & User Flows

**Duration:** 2 weeks (parallel with Phase 1 implementation)
**Deliverables:**

### Dashboard Information Architecture
- [ ] Sitemap (all pages, hierarchy)
- [ ] Navigation structure (sidebar, top nav)
- [ ] Page inventory (list all pages needed)

### User Flows
- [ ] Admin login flow
- [ ] Location switching flow
- [ ] Gate configuration flow (new gate detected → configure)
- [ ] Rate management flow (create/edit rates)
- [ ] Shift config flow
- [ ] User management flow (create user, assign role)
- [ ] Report viewing flow (select report, filter, export)
- [ ] Alert response flow (see alert → investigate → resolve)

### Gate App User Flows
- [ ] Entry flow (driver perspective)
  - Vehicle detected → button press → ticket dispensed → gate opens
- [ ] Exit flow (driver perspective)
  - QR scan → fee displayed → payment → gate opens
  - Payment failed → message → retry or alert
- [ ] Alert flow (driver presses button → staff notified)

### Server Room App User Flows
- [ ] Staff monitoring flow (see gate status, respond to alerts)
- [ ] Manual refresh flow (refresh configs, refresh data)
- [ ] Offline mode flow (system offline → staff notified → manual override)

**Stakeholders:** UX designer, product manager, engineering lead

---

## Week 3-4: Wireframes & Low-Fidelity Mockups

**Duration:** 2 weeks (parallel with Phase 2-3 implementation)
**Deliverables:**

### Dashboard Wireframes
- [ ] Login page
- [ ] Main dashboard (live monitoring overview)
- [ ] Gate status page (list of gates, status indicators)
- [ ] Gate configuration modal (new gate setup)
- [ ] Location selector (dropdown, grouped by city)
- [ ] Rate management page (list, create/edit form)
- [ ] Shift config page
- [ ] User management page (list, create/edit form)
- [ ] Reports pages (daily revenue, occupancy, transactions)
- [ ] Alert history page

### Gate App Wireframes (Driver-Facing Monitor)
- [ ] Entry screen (idle state: "Press button for ticket")
- [ ] Entry processing screen ("Dispensing ticket...")
- [ ] Entry success screen ("Ticket dispensed, gate opening")
- [ ] Exit screen (idle state: "Scan your ticket")
- [ ] Exit fee display screen (fee amount, check-in time, duration, "Please tap your card")
- [ ] Payment processing screen ("Processing payment...")
- [ ] Payment success screen ("Payment successful, gate opening")
- [ ] Payment failed screen ("Insufficient balance, please topup")
- [ ] Error screen ("Out of service, please press alert button")
- [ ] Receipt screen (optional: check-in/out time, fee, vehicle type)

### Server Room App Wireframes (Staff Dashboard)
- [ ] Main monitoring screen (gate status grid, alerts)
- [ ] Alert detail screen (gate ID, issue type, timestamp)
- [ ] Offline mode screen (system status, manual override options)

### Hardware Interaction Mockups
- [ ] Driver-facing monitor placement (height, angle, visibility)
- [ ] Ticket button placement (height, accessibility)
- [ ] QR scanner placement (height, angle)
- [ ] Payment terminal placement (height, reach)
- [ ] Alert button placement (height, visibility)
- [ ] Staff monitor placement (server room, visibility)

**Stakeholders:** UX designer, product manager, engineering lead, AMB staff (feedback)

---

## Week 5-6: Visual Design & High-Fidelity Mockups

**Duration:** 2 weeks (parallel with Phase 4 implementation)
**Deliverables:**

### Dashboard Visual Design
- [ ] Design system (colors, typography, spacing, components)
- [ ] Component library (buttons, forms, tables, cards, modals)
- [ ] High-fidelity mockups (all pages)
- [ ] Responsive design (desktop, tablet, mobile)
- [ ] Dark mode (optional, if requested)
- [ ] Accessibility audit (color contrast, font size, keyboard navigation)

### Gate App Visual Design (Driver-Facing Monitor)
- [ ] Visual design system (large fonts, high contrast, simple icons)
- [ ] High-fidelity mockups (all screens)
- [ ] Animation/transition design (smooth state changes)
- [ ] Multi-language support (Indonesian primary, English optional)
- [ ] Accessibility audit (font size for drivers, color blindness)

### Server Room App Visual Design (Staff Dashboard)
- [ ] Visual design system (consistent with dashboard)
- [ ] High-fidelity mockups (all screens)
- [ ] Alert notification design (visual + audio)

### Interactive Prototypes
- [ ] Dashboard prototype (clickable, Figma or similar)
- [ ] Gate app prototype (driver flow simulation)
- [ ] Server room app prototype (staff flow simulation)

### Design Documentation
- [ ] Design specs (measurements, colors, fonts, spacing)
- [ ] Component documentation (usage guidelines)
- [ ] Interaction documentation (animations, transitions)
- [ ] Asset export (icons, images)

**Stakeholders:** UX designer, UI designer, product manager, engineering lead, AMB staff (feedback)

---

## Week 7: Design System & Component Library

**Duration:** 1 week (parallel with Phase 6 implementation)
**Deliverables:**

### Design System Finalization
- [ ] Color palette (primary, secondary, accent, semantic colors)
- [ ] Typography scale (headings, body, captions)
- [ ] Spacing system (4px, 8px, 16px, 24px, 32px)
- [ ] Icon library (custom icons or icon set)
- [ ] Component library (all reusable components)

### Component Library (for developers)
- [ ] React components (dashboard)
- [ ] Gate app components (driver-facing UI)
- [ ] Server room app components (staff UI)
- [ ] Storybook or similar (component documentation)

### Design Tokens
- [ ] Design tokens (JSON/CSS variables for colors, fonts, spacing)
- [ ] Theme configuration (light/dark mode, if applicable)

**Stakeholders:** UI designer, frontend developers

---

## Week 8: Usability Testing & Iteration

**Duration:** 1 week (parallel with Phase 7 implementation)
**Deliverables:**

### Usability Testing
- [ ] Dashboard usability test (5-7 AMB admin users)
  - Tasks: login, switch location, configure gate, view report
  - Measure: time on task, error rate, satisfaction
- [ ] Gate app usability test (5-7 drivers)
  - Tasks: enter parking, exit parking, handle payment failure
  - Measure: time on task, error rate, satisfaction
- [ ] Server room app usability test (3-5 staff)
  - Tasks: monitor gates, respond to alert, handle offline mode
  - Measure: time on task, error rate, satisfaction

### Iteration
- [ ] Analyze test results (identify pain points)
- [ ] Prioritize fixes (critical, high, medium, low)
- [ ] Update designs (fix critical + high priority issues)
- [ ] Re-test (if major changes)

### Accessibility Audit
- [ ] WCAG 2.1 AA compliance check
- [ ] Color contrast audit
- [ ] Keyboard navigation test
- [ ] Screen reader test (dashboard)

**Stakeholders:** UX researcher, UX designer, product manager, AMB users

---

## Week 9-10: Design Handoff & Support

**Duration:** 2 weeks (parallel with Phase 8 implementation)
**Deliverables:**

### Design Handoff
- [ ] Final design files (Figma, Sketch, or similar)
- [ ] Design specs (all measurements, colors, fonts)
- [ ] Asset export (icons, images, fonts)
- [ ] Component library (Storybook or similar)
- [ ] Design documentation (usage guidelines)

### Developer Support
- [ ] Design QA (review implementation, flag discrepancies)
- [ ] Bug fixes (design-related bugs)
- [ ] Design tweaks (minor adjustments based on implementation constraints)

### Pilot Support
- [ ] Observe pilot deployment (Week 9-10)
- [ ] Collect user feedback (AMB staff, drivers)
- [ ] Identify design improvements
- [ ] Plan iteration for next release

**Stakeholders:** UX designer, UI designer, frontend developers, product manager

---

## Design Timeline Summary

| Week | Design Activity | Parallel With |
|------|----------------|---------------|
| **Week -2 to -1** | Research & Discovery | Pre-implementation |
| **Week 1-2** | Information Architecture & User Flows | Phase 1 (Foundation) |
| **Week 3-4** | Wireframes & Low-Fidelity Mockups | Phase 2-3 (Entry/Exit Flow) |
| **Week 5-6** | Visual Design & High-Fidelity Mockups | Phase 4 (Dashboard) |
| **Week 7** | Design System & Component Library | Phase 5 (Offline & Sync) |
| **Week 8** | Usability Testing & Iteration | Phase 6 (Monitoring) |
| **Week 9-10** | Design Handoff & Support | Phase 7-8 (Testing & Pilot) |

---

## Key Design Milestones

| Milestone | Week | Deliverable |
|-----------|------|-------------|
| **M1: Research Complete** | Week -1 | User research report, competitive analysis |
| **M2: User Flows Approved** | Week 2 | Sitemap, user flows (all apps) |
| **M3: Wireframes Approved** | Week 4 | Low-fidelity mockups (all apps) |
| **M4: Visual Design Approved** | Week 6 | High-fidelity mockups, interactive prototypes |
| **M5: Design System Complete** | Week 7 | Component library, design tokens |
| **M6: Usability Test Complete** | Week 8 | Test results, iteration plan |
| **M7: Design Handoff Complete** | Week 10 | Final designs, component library, docs |

---

## Design Team Roles

| Role | Responsibility | Time Commitment |
|------|---------------|-----------------|
| **UX Researcher** | User interviews, usability testing | Week -2 to -1, Week 8 |
| **UX Designer** | Information architecture, user flows, wireframes | Week 1-4, Week 8 |
| **UI Designer** | Visual design, design system, component library | Week 5-7, Week 9-10 |
| **Product Manager** | Requirements, stakeholder management, prioritization | Full timeline |

**Total design effort:** ~12 weeks (2 weeks pre-implementation + 10 weeks parallel)

---

## Design Tools

| Tool | Purpose |
|------|---------|
| **Figma** | Wireframes, mockups, prototypes, design system |
| **FigJam** | User flows, information architecture |
| **Maze or UserTesting** | Remote usability testing |
| **Storybook** | Component library documentation |
| **Notion** | Design documentation, research reports |

---

## Design Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Limited access to AMB staff for research | Start research early (Week -2), schedule interviews in advance |
| Hardware constraints unknown | Get hardware specs early (Week -2), design flexible layouts |
| Usability issues discovered late | Test wireframes (Week 4) + high-fidelity (Week 8), iterate early |
| Design-development handoff issues | Use design tokens, component library, regular sync meetings |
| Scope creep (new features requested) | Freeze scope at Week 4 (wireframes approved), defer to post-MVP |

---

## Post-MVP Design Roadmap

**Month 3:** Iterate based on pilot feedback (Week 9-10 findings)
**Month 4:** Advanced features design (receipt printing, multi-language)
**Month 5:** Performance optimization (animations, loading states)
**Month 6:** Accessibility improvements (WCAG AAA, if needed)

---

*End of Design Timeline*
