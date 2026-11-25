# Search syntax

DPMA Connect Plus search (`SearchPatents` / `SearchPatentsParsed`,
`SearchDesigns` / `SearchDesignsParsed`, `SearchTrademarks` /
`SearchTrademarksParsed`) accepts the **DPMAregister "Expertenrecherche" (expert
search) logical query language** - a proprietary `FIELD=value` expression with
Boolean and comparison operators. Each method takes the expression as a plain
string and posts it to the corresponding service. This document is the
query-syntax reference so the README stays uncluttered (the README links here).

**Official source:**
- DPMAconnect (Connect Plus) REST API spec ("Schnittstellenbeschreibung
  DPMAconnectPlus"):
  <https://www.dpma.de/docs/recherche/dienste/schnittstellenbeschreibungdpmaconnectplus.pdf>
- DPMAregister expert-search help, "Recherchefelder" (field codes) pages, one
  per service:
  - Patents/utility models: <https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/patgbm/index.html>
  - Trademarks: <https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/marken/index.html>
  - Designs (Geschmacksmuster): <https://register.dpma.de/register/htdocs/prod/de/hilfe/recherchefelder/gsm/index.html>
- Expert-search entry point: <https://register.dpma.de/DPMAregister/pat/experte>

The field-code tables, operators, and wildcards below are transcribed from the
library's own field definitions in `query/fields.go` and the tokenizer/validator
in `query/query.go`, which were derived from the official DPMAregister help cited
above. The README's "Search query syntax" section is a consistent cross-check.
Last checked: 2026-06-12.

**Verification:** the patent fields were confirmed against the live DPMA Connect
Plus API on 2026-06-12 via `SearchPatentsParsed`:

- `TI=Batterie AND INH=Bosch` -> 998 hits
- `TI=Elektrofahrzeug` -> 4075 hits

Individual field codes below come from the library's own definitions (derived
from the official DPMAregister help) and are **not** all individually
live-verified; treat any single code as authoritative per the help pages but
unconfirmed against the live API unless noted. Note also: some accounts return
`No permission of user=sip` (a known per-account permission limit) - if you hit
that, the query syntax is fine but your account lacks access to that service.

## Searchable fields

Field codes are **per service**. The same code can mean different things in
different services (e.g. `ET` is the design registration date in the design
service, and the utility-model/topography registration date in the patent
service). Use the table for the service you are querying.

### Patent / utility model fields (`SearchPatents`)

| Field | Meaning |
|---|---|
| `AKZ` | file number / publication number (Aktenzeichen/Veroeffentlichungsnummer) |
| `PN` | publication number, DEPN + EPWOPN combined (Veroeffentlichungsnummer) |
| `PNN` | publication number without country/kind code |
| `INH` | applicant / proprietor (Anmelder/Inhaber) |
| `IN` | inventor (Erfinder) |
| `AT` | filing date (Anmeldetag) |
| `DAT` | effective filing date for Germany (Anmeldetag DE) |
| `TI` | title / designation (Bezeichnung/Titel) |
| `AB` | abstract text |
| `FT` | fee due date (Faelligkeitstag) |
| `FG` | fee due reason (Faelligkeit Grund) |
| `FOEZ` | public funding identifier (Foerderkennzeichen) |
| `IC` | IPC classification, main + secondary (IPC-Klasse) |
| `ICM` | IPC main classification (IPC-Hauptklasse) |
| `ICMV` | IPC main classification version |
| `ICS` | IPC secondary classification (IPC-Nebenklasse) |
| `ICSV` | IPC secondary classification version |
| `LIZ` | license declaration (Lizenz) |
| `PUB` | publication date (Publikationstag) |
| `SART` | protection right type: Patent, Gebrauchsmuster, etc. (Schutzrechtsart) |
| `ST` | current status of the protection right (Status) |
| `EREGT` | date first entered in DPMAregister |
| `REGT` | date of last update in DPMAregister |
| `VSTT` | procedure status date (Verfahrensstandtag) |
| `VST` | procedure status, use in curly braces (Verfahrensstand) |
| `PET` | publication date of grant/registration (Veroeffentlichungstag der Erteilung/Eintragung) - this is the patent **grant** publication field |
| `VTR` | representative / patent attorney (Vertreter) |
| `EVT` | date of first publication (Erstveroeffentlichungstag) |
| `EAT` | effective filing date for EP patent (Anmeldetag EP) |
| `WAT` | effective filing date for PCT application (Anmeldetag WO) |
| `PRC` | foreign priority country |
| `PRNA` | foreign priority application number |
| `PRDA` | foreign priority date |
| `APR` | exhibition priority name (Ausstellungsprioritaet) |
| `APRD` | exhibition priority date |
| `EDS` | designated contracting states EP |
| `DS` | designated states, EDS + WDS combined |
| `WDS` | designated states WO/PCT |
| `CT` | citations / prior art references (Entgegenhaltungen) |
| `CTNP` | non-patent literature citations |
| `DPRN` | design priority application number |
| `DPRC` | design priority country |
| `DPRD` | design priority date |
| `PRN` | domestic priority application number (Inlaendische Prioritaet) |
| `PRD` | domestic priority date |
| `TAAKZ` | divisional application file number (Teilung/Ausscheidung aus) |
| `TIAKZ` | split-off application file number (Teilung/Ausscheidung in) |
| `DEPN` | published DE document numbers |
| `EPWOPN` | published EP/WO document numbers |
| `PC` | publication country code |
| `PCOD` | publication kind code (Schriftartencode) |
| `EPN` | EP publication number |
| `WPN` | WO/PCT publication number |
| `LANG` | publication language, ELANG + WLANG combined |
| `ELANG` | EP publication language |
| `WLANG` | WO publication language |
| `AAAKZ` | branched-off from file number (Abzweigung, utility model) |
| `EAKZ` | EP file number linked to utility model |
| `ET` | registration date - **utility model / topography** (Eintragungstag); NOT a patent grant date (use `PET` for grant publication) |
| `GBMNAT` | PCT national phase entry date for utility model |
| `NVWT` | date of first non-confidential use (topography) |
| `TIEZ` | product designation, SPC (Bezeichnung des Erzeugnisses) |
| `TIGZ` | basic patent title, SPC (Bezeichnung des Grundpatents) |
| `GEC` | EU marketing authorization country (SPC) |
| `GEN` | EU marketing authorization number (SPC) |
| `GES` | EU marketing authorization authority (SPC) |
| `GET` | EU marketing authorization date (SPC) |
| `GIN` | domestic marketing authorization number (SPC) |
| `GIS` | domestic marketing authorization authority (SPC) |
| `GIT` | domestic marketing authorization date (SPC) |
| `GAKZ` | basic patent file number (SPC) |
| `GEAKZ` | basic patent EP file number (SPC) |
| `LB` | SPC term start date (Laufzeit Beginn) |
| `LE` | SPC term end date (Laufzeit Ende) |
| `OT` | SPC application publication date |
| `VZ` | SPC use: Arzneimittel/Pflanzenschutzmittel (Verwendung des Zertifikats) |
| `PART` | procedure type for patent gazette (Verfahrensart Patentblatt) |
| `VT` | procedure status month, format `Mnn-yyyy` (Verfahrensstand-Monat) |
| `HN` | gazette issue number (Heftnummer) |
| `PJ` | publication year (Publikationsjahr) |

### Design fields (`SearchDesigns`, Geschmacksmuster)

| Field | Meaning |
|---|---|
| `AKZ` | file number / design number (Aktenzeichen/Designnummer) |
| `INH` | proprietor (Inhaber) |
| `ENTW` | designer (Entwerfer) |
| `ERZ` | product(s) (Erzeugnis(se)) |
| `TI` | title / designation (Bezeichnung/Titel) |
| `WKL` | Locarno class (Warenklasse) |
| `RN` | registration number (Registernummer) |
| `DNR` | design number (Designnummer) |
| `AT` | filing date (Anmeldetag) |
| `ET` | registration date (Eintragungstag) |
| `ST` | current status (Status) |
| `VTR` | representative (Vertreter) |
| `LIZ` | license declaration (Lizenz) |
| `BA` | stock type: aktiv / nicht-aktiv (Bestandsart) |
| `DB` | database: DE / EM (Datenbestand) |
| `PUB` | publication date (Publikationstag) |
| `EREGT` | date first entered in DPMAregister |
| `REGT` | date of last update in DPMAregister |
| `VART` | procedure type (Verfahrensart) - curly braces |
| `VSTT` | procedure status date (Verfahrensstandtag) - curly braces |
| `VST` | procedure status (Verfahrensstand) - curly braces |

### Trademark fields (`SearchTrademarks`, Marken)

| Field | Meaning |
|---|---|
| `MARKE` | trademark text (Marke) |
| `md` | trademark text (alias for `MARKE`) |
| `INH` | applicant / proprietor (Anmelder/Inhaber) |
| `AT` | filing date (Anmeldetag) |
| `RN` | registration number / file number (Registernummer/Aktenzeichen) |
| `BA` | status type: eingetragen, eintragung-nicht-moeglich, etc. (Bestandsart) |
| `DB` | database: DE / EM / IR (Datenbestand) |
| `KL` | Nice class(es) (Klasse(n)) |
| `BKL` | Vienna / image class(es) (Bildklasse(n)) |
| `MF` | trademark form: Wort-/Bildmarke etc. (Markenform) |
| `LIZ` | license declaration (Lizenzerklaerung) |
| `VTR` | representative (Vertreter) |
| `ST` | current status (Status) |
| `VART` | procedure type (Verfahrensart) - curly braces |
| `VSTT` | procedure status date (Verfahrensstandtag) - curly braces |
| `VST` | procedure status (Verfahrensstand) - curly braces |

The library also accepts a unary `exists FIELD` form (e.g. `exists INH`) to match
records that have any value in a field.

## Operators

- **Boolean:** `AND` / `UND`, `OR` / `ODER`, `NOT` / `NICHT` - English and German
  forms are both accepted (case-insensitive lower-case variants too). `NOT`/`NICHT`
  means "and not" and may not start an expression. Adjacent or dangling Boolean
  operators are rejected by the client-side validator.
- **Comparison:** `=`, `>=`, `<=`, `>`, `<` (e.g. `AT>=01.01.2024`). Every
  `FIELD <op> value` needs a value on the right-hand side.
- **Wildcards / truncation:** `?` = any number of characters including none;
  `!` = exactly one character; `#` = one character or none. (e.g.
  `IC=H02K?`, `MARKE=?brain?`.)
- **Grouping:** parentheses `( ... )` for Boolean grouping (must be balanced).
- **Phrase / quoting:** double quotes `"..."` for multi-word or special-character
  values, e.g. `INH="Robert Bosch"`.
- **Procedure data:** curly braces `{ ... }` bind procedure-status fields together,
  e.g. `{VST=pub-offenlegungsschrift UND VSTT=05.01.2011}` (braces must be balanced).

## Examples

```text
TI=Elektrofahrzeug
TI=Batterie AND INH=Bosch
TI=Batterie UND INH=Bosch
(TI=Motor OR TI=Antrieb) AND IC=H02K?
INH="Robert Bosch" AND AT>=01.01.2024
MARKE=?brain?
exists INH
{VST=pub-offenlegungsschrift UND VSTT=05.01.2011}
```

(`TI=Batterie AND INH=Bosch` and `TI=Elektrofahrzeug` were run live on
2026-06-12 and returned hits. On 2026-09-09 further live probes confirmed
`VTR=` on all three services, the `?` wildcard (`TI=Elektro?`), umlauts
(`VTR=Gruenecker` with the u-umlaut percent-encoded in the URL), and
`exists FOEZ`.)

## Notes and limitations

- **Date format:** `dd.mm.yyyy` (e.g. `01.01.2024`). The procedure-status month
  field `VT` uses `Mnn-yyyy`; publication weeks for bulk downloads (a separate API
  area) use `YYYYWW`.
- **Per-service fields:** field validity is service-specific. The client-side
  validator (`ValidatePatentQuery`, `ValidateDesignQuery`, `ValidateTrademarkQuery`,
  or `query.ParseQuery` with a `Service`) rejects fields not defined for that
  service.
- **Case:** field codes are upper-case (the `md` trademark alias is lower-case);
  Boolean operators are case-insensitive.
- **Account permissions:** a `No permission of user=sip` response is a known
  per-account access limit, not a syntax problem - the account simply lacks
  permission for that service or operation.
- **Result cap:** every search service caps the returned hit list at 10000
  records. A broader query still succeeds (HTTP 200) but the root element then
  carries `maxHitsReached="true"` plus the notice "Result list is limited - Not
  all results are returned" (patent/design roots carry it as `Message_DE`/
  `Message_EN` attributes; on all services the `<Counter><DocumentHits>` value
  is replaced by the notice text and only `<DatabaseHits>` keeps the
  register-wide total). The parsed results expose this as `MaxHitsReached`.
  There is no pagination; narrow the query to reach hits beyond the cap.
- **Unverified:** only the queries listed above were run against the live API.
  Individual field codes are sourced from the library's definitions (per the
  official DPMAregister help) and have not each been live-confirmed.
