# ASDU Parsed Message Refactor Plan

Goal: replace destructive on-demand parsing (`ASDU.Get*`) and multi-method handler interfaces with a single handler that receives a parsed, type-assertable message. Parsing happens once on receive.

## Progress Legend
- [ ] not started
- [~] in progress
- [x] done

## 1) Inventory and Scope Confirmation
- [x] Map all inbound parse call sites and handler dispatchers (client/server)
- [x] Identify all `Get*` decoders to be replaced or wrapped
- [x] Note tests relying on idempotent `Get*` behavior

## 2) Parsed Message Model Design
- [x] Define `asdu.Message` interface and `Header` carrier struct
- [x] Decide grouping strategy (family-level structs as sketched)
- [x] Decide how raw payload is retained for unknown/private TypeIDs
- [x] Decide whether to keep `*ASDU` around as transport-only (no parsing)

## 3) Parser Implementation (Non-Destructive)
- [x] Create a small read cursor over a byte slice (`infoObj` + offset)
- [x] Implement decoders using the cursor (mirror existing decode methods)
- [x] Add `ParseASDU(*ASDU) (Message, error)` entry point
- [x] Return `UnknownMsg` for unsupported/unknown TypeIDs

## 4) Handler API Refactor
- [x] Replace `ClientHandlerInterface` and `ServerHandlerInterface` with a single handler method (e.g., `Handle(asdu.Connect, asdu.Message) error`)
- [x] Update `cs104/client.go` handler loop to parse once and call new handler
- [x] Update `cs104/server_session.go` handler loop similarly
- [x] Port server-side validation/reply logic to use parsed messages (e.g., `InterrogationCmdMsg`)

## 5) Deprecate or Rework `Get*` APIs
- [x] Decide: remove `Get*` or keep as thin wrappers over `ParseASDU`
- [x] Remove `Get*` APIs in favor of parsed messages
- [x] Update `ASDU.String` and `ASDU.MarshalJSON` to use parsed messages

## 6) Examples and Tests
- [x] Update `_examples/**` to use type assertions on parsed messages
- [x] Remove or rewrite idempotent `Get*` tests
- [x] Add parser tests per message family (round-trip or known vectors)
- [x] Add handler dispatch tests (type assertion paths)

## 7) Documentation and Migration Notes
- [x] Update `README.md` with new handler interface
- [x] Add migration notes for former handler methods and `Get*` usage

---

### Implementation Checklist (as work begins)
- [x] Implement `asdu/parse.go` (new) with `ParseASDU` and message structs
- [x] Update `cs104/interface.go` to new handler interface
- [x] Update `cs104/client.go` to parse and call handler
- [x] Update `cs104/server_session.go` to parse and call handler
- [x] Update `asdu/asdu.go` `String`/`MarshalJSON`
- [x] Update examples and tests

Notes:
- Keep changes ASCII-only unless files already use Unicode.
- Preserve any unrelated user changes in the worktree.
