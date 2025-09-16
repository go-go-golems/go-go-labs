Here’s a tight research brief with concrete answers, citations, and a pragmatic path forward.

# Research Brief: Embedding Stable Metadata in Google Forms Items

## Summary (1–2 pages)

### 1) Can clients set `itemId` / `questionId` on creation?

* **`Item.itemId`** — Yes, the Forms REST reference explicitly says: “On creation, it can be provided but the ID must not be already used in the form. If not provided, a new ID is assigned.” ([Google for Developers][1])
* **`Question.questionId`** — The same page marks it “Read only… On creation, it can be provided but the ID must not be already used in the form.” (Yes, the text is paradoxical—“read only” yet “can be provided on creation”—but this is how the official doc currently reads.) ([Google for Developers][1])
* **Batch update behavior** — For `UpdateItemRequest`, “item and question IDs are used if they are provided (and are in the field mask). If an ID is blank (and in the field mask) a new ID is generated.” This confirms IDs are server-validated and can be respected when present. ([Google for Developers][2])

**Reality check on formats:** Community testing indicates **IDs must be 8-char hex within `00000000`–`7fffffff`**; non-hex strings (e.g., Base64 like `uhoh__…`) tend to fail with `400 Invalid ID`. This is **not documented** by Google, but it’s the most reliable public guidance from reproducible tests. ([Stack Overflow][3])

### 2) Alternate fields for per-item metadata (without affecting respondents)

There are **no hidden, item-scoped “metadata” fields** in the Forms API schema. All obvious string fields on items/questions (title, description, option values, alt text) are **user-visible**. See the schema for `Item`, `Question`, `Option`, etc. ([Google for Developers][1])

A robust alternative is to store your metadata **outside the item body** and link by stable IDs:

* **Drive file `appProperties` / `properties`** on the Form file (via Drive API). These are key/value pairs attached to the Form in Drive; **`appProperties` are private to your app** and are searchable. Limits: up to **100 custom properties per file**, **≤30 private properties per app**, **124 bytes combined key+value per property**. ([Google for Developers][4])

### 3) Recommended patterns (Google/community)

* **Use the server’s (or valid client-provided) IDs as the join key** and keep your DSL mapping outside of the user-visible text. Google’s `CreateItemResponse` returns both `itemId` and `questionId`—capture and persist them. ([Google for Developers][2])
* **Attach a compact mapping to the Form file** using Drive API `appProperties` (e.g., store a hashed/segmented mapping or a pointer to your own datastore). This pattern (Drive custom properties for app-level metadata) is explicitly supported and documented by Google. ([Google for Developers][4])
* Community guidance warns against stuffing IDs into titles/descriptions/options because they’re respondent-visible and brittle. (SO threads & docs above.) ([Stack Overflow][3])

### 4) Limits & caveats

* **Item/Question IDs**: Google docs don’t list a format, but practical success cases use **8-char hex**. Using other alphabets (Base64) often triggers `400 Invalid ID`. Treat hex as the de-facto constraint. ([Stack Overflow][3])
* **Drive `appProperties`**:

  * Per-file: **≤100** custom properties (all sources).
  * Per app: **≤30** private properties.
  * Per property: **≤124 bytes** for `key + value` UTF-8. Consider sharding or storing a pointer to an external record if your mapping is large. ([Google for Developers][4])
* **User-visible fields** (`Item.title`, `Item.description`, `Option.value`, `Image.altText`) will be seen by respondents; use only if you *want* the metadata visible. ([Google for Developers][1])

### 5) Creation vs. update flows (writing metadata post-creation)

* **Creation**: You may *provide* `itemId` (and per docs, `questionId`) on create; otherwise Google generates them. Capture returned IDs from `CreateItemResponse`. ([Google for Developers][2])
* **Update**: `UpdateItemRequest` respects IDs when included and in the `updateMask`. If you leave an ID blank *and* it’s in the mask, a **new ID is generated**. In practice, do **not** churn IDs after you start collecting responses, since responses are keyed by `questionId`. ([Google for Developers][2])

### 6) If no field fits: external linkage options

* **Drive `appProperties`** on the Form file (best blend of “near the form” + hidden + queryable). ([Google for Developers][4])
* **External DB** keyed by `formId` + (`itemId`/`questionId`) with an optional **pointer stored in `appProperties`** to facilitate discovery. ([Google for Developers][4])
* **Add-on** approach (Apps Script): possible but unnecessary if you already control the API caller; add-ons don’t add hidden per-item storage. (General reference to Apps Script/Forms docs.) ([Google for Developers][5])

---

## Contrast Table: Candidate Places to Store Metadata

| Field / Place                             |                                                    Max length (official) | Visible to respondents? | Pros                                                                              | Cons                                                                                     |
| ----------------------------------------- | -----------------------------------------------------------------------: | ----------------------- | --------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `Item.itemId`                             | Not documented; community success with **8-hex** (`00000000`–`7fffffff`) | No                      | Stable, designed as identifier; can be provided on create; round-trippable in API | Format constraints are undocumented; collisions/validation to manage; not human-readable |
| `Question.questionId`                     |                                                           Not documented | No                      | Stable key used in responses; returned by API                                     | “Read only” label in docs is confusing; don’t try to mutate post-creation                |
| `Item.title`                              |                                                            Not specified | **Yes**                 | Easy to set/update                                                                | User-visible; brittle if content changes; localization issues                            |
| `Item.description`                        |                                                            Not specified | **Yes**                 | Roomy text field                                                                  | User-visible; formatting/escaping visible; easy to accidentally alter                    |
| `Option.value`                            |                                                            Not specified | **Yes**                 | Already round-trips in answers                                                    | Must be visible choices; changes affect data integrity                                   |
| `Image.altText`                           |                                                            Not specified | **Yes**                 | Can hide in media                                                                 | Still visible to screen readers/hover; awkward                                           |
| **Drive `appProperties`** (Form file)     |      **124 bytes (key+value)**; **≤30 private/app**, **≤100 total/file** | **No**                  | Hidden, app-scoped, searchable, ideal for pointers/indexes                        | Tight size; needs Drive scope; per-file cap—may need sharding                            |
| External DB (pointer via `appProperties`) |                                                                      N/A | No                      | Unlimited structure; versioning                                                   | Requires infra; ensure pointer durability                                                |

Citations: IDs & schema ([Google for Developers][1]); batch update semantics & create response ([Google for Developers][2]); hex-ID practice ([Stack Overflow][3]); Drive properties limits & usage ([Google for Developers][4]).

---

## Clear Recommendation

**Use IDs, not visible text.** Adopt a two-layer approach:

1. **Deterministic, valid item IDs at creation time (preferred):**

   * Derive an **8-char lowercase hex** from your DSL ID (e.g., `lower(hex(sha1(dsl_id)))[0:8]`).
   * Provide it as `Item.itemId` in `createItem` requests; let Google keep it if valid.
   * Capture `CreateItemResponse.itemId` and `questionId[]` for each item for your mapping table.
     Rationale: Matches the official “can be provided on create” rule; avoids user-visible hacks; gives you round-trip stability. ([Google for Developers][1])
     Field-format caveat: The best-available public evidence says IDs must be 8-char hex in range `00000000`–`7fffffff`. Don’t use Base64/underscores. ([Stack Overflow][3])

2. **Persist a compact mapping on the Form itself with Drive `appProperties`:**

   * Store **chunks** like `map_00`…`map_1F` (up to 30 private props) holding compressed/JSON-minified pairs (`itemId` → `dsl_id`, optionally `questionId` too).
   * Each key/value must fit **≤124 bytes**; so shard or store a **URL/ID pointer** to your external datastore if the mapping is large. ([Google for Developers][4])
   * This keeps the binding discoverable from just the `formId` with Drive API, without exposing anything to respondents. ([Google for Developers][4])

**Implementation tips**

* **Create flow**:

  * For each DSL node, compute `itemId_hex8`; include it in `CreateItemRequest.item.itemId`.
  * After `batchUpdate`, read `CreateItemResponse` and persist `{dsl_id, itemId, questionId}`. ([Google for Developers][2])
* **Update flow**:

  * Retrieve the current form (`forms.get`) and your mapping (from `appProperties` and/or DB).
  * Use `UpdateItemRequest` with `updateMask` limited to the fields you’re changing; **don’t blank IDs** unless you intend to create new ones. ([Google for Developers][2])
* **When IDs can’t be pre-set** (e.g., you’d rather let Google generate):

  * Accept server IDs, then immediately write your `{itemId,questionId}→dsl_id` mapping into `appProperties` (or your DB) as part of the same operation cycle. ([Google for Developers][4])
* **Avoid** stuffing metadata into `title/description/option.value`—it’s visible and fragile. ([Google for Developers][1])

---

## Direct Answers to Your Questions (with sources)

1. **Client-assigned IDs?**

   * `Item.itemId`: **Allowed on creation** if unique. ([Google for Developers][1])
   * `Question.questionId`: Docs say “Read only” but also “can be provided on creation” (treat as server-validated; rely on the returned IDs). ([Google for Developers][1])
   * `UpdateItemRequest` honors provided IDs and generates if blank (within mask). ([Google for Developers][2])

2. **Alternate fields for structured metadata?**

   * None that are hidden and item-scoped. All in-schema text is visible. Use **Drive `appProperties`** on the Form file or an external store keyed by `itemId`/`questionId`. ([Google for Developers][1])

3. **Recommended patterns?**

   * **Capture IDs and link externally**; store mappings in **Drive `appProperties`** or your DB. ([Google for Developers][2])
   * Community practice for client-provided IDs = **8-hex**. ([Stack Overflow][3])

4. **Limits/caveats?**

   * ID format: use 8-hex; Base64/underscores → `400 Invalid ID`. (Community evidence) ([Stack Overflow][3])
   * `appProperties`: **124 bytes (key+value)**, **≤30 private per app**, **≤100 total per file**. ([Google for Developers][4])

5. **Creation vs update differences?**

   * You can **write metadata post-creation** by updating your external mapping and/or `appProperties`.
   * `UpdateItemRequest` uses provided IDs; blank IDs in the mask generate new ones—**don’t** do this after responses begin. ([Google for Developers][2])

6. **External solutions?**

   * **Drive `appProperties`** (preferred “on-the-form” storage). ([Google for Developers][4])
   * **External DB** keyed by `formId` + IDs, optionally linked via an `appProperties` pointer. ([Google for Developers][6])

---

## Recommendation You Can Implement Now

* **Stop trying Base64 IDs.** Switch to **8-char hex** derived deterministically from your DSL IDs (e.g., `sha1(dsl_id)[:8]` lowercased). Provide as `Item.itemId` during **create**; capture the **returned** `itemId`/`questionId` as the source of truth. ([Google for Developers][1])
* **Persist mappings on the Form** using **Drive `appProperties`**, sharded to respect the **124-byte** limit, or store a pointer to your datastore if the mapping is large. ([Google for Developers][4])
* **Never use respondent-visible fields** for hidden metadata. ([Google for Developers][1])

If you want, I can sketch a tiny, copy-pasteable create/update flow (Go/JS) that: (a) builds hex IDs, (b) calls `forms.batchUpdate`, (c) captures `CreateItemResponse`, and (d) writes a compact mapping into `appProperties`.

[1]: https://developers.google.com/workspace/forms/api/reference/rest/v1/forms "REST Resource: forms  |  Google Forms  |  Google for Developers"
[2]: https://developers.google.com/workspace/forms/api/reference/rest/v1/forms/batchUpdate "Method: forms.batchUpdate  |  Google Forms  |  Google for Developers"
[3]: https://stackoverflow.com/questions/73560879/generate-itemid-when-batchupdating-with-forms-api "node.js - Generate itemId when batchUpdating with Forms API - Stack Overflow"
[4]: https://developers.google.com/workspace/drive/api/guides/properties?utm_source=chatgpt.com "Add custom file properties | Google Drive"
[5]: https://developers.google.com/apps-script/reference/forms/form?utm_source=chatgpt.com "Class Form | Apps Script"
[6]: https://developers.google.com/workspace/drive/api/guides/search-files?utm_source=chatgpt.com "Search for files and folders | Google Drive"

