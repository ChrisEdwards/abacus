# Architecture Decisions

This document captures key design decisions for the Abacus project.

---

## 001: Never auto-create project-level config

**Date:** 2025-12-06

**Context:** Abacus supports both global config (`~/.abacus/config.yaml`) and project-level config (`.abacus/config.yaml`). When persisting settings like theme selection, we need to decide where to write.

**Decision:** Never auto-create a `.abacus/` folder or config file in the user's project directory.

**Rationale:**
- Avoid polluting user repos with unexpected files
- Project-level config could get committed to git by accident
- No surprises - user explicitly opts in by creating `.abacus/config.yaml`

**Behavior:**
- If `.abacus/config.yaml` exists in project → use it (user opted in)
- Otherwise → use global `~/.abacus/config.yaml`

**Related beads:** ab-x4bw, ab-3d7u
