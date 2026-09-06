VERDICT: CHANGES_REQUESTED

## Zusammenfassung

Der Feature-Flag-Service ist als reines Go-Backend ohne Endnutzer-UI einzuordnen. Impressums-, Cookie-, Consent- und Web-Accessibility-Pflichten sind daher nicht anwendbar. Der sichtbare Code ist in weiten Teilen datenschutzfreundlich und sicher umgesetzt: Die Logging-Middleware protokolliert nachweislich keine Query-Strings oder User-IDs, die `recover`-Middleware vermeidet Stacktraces und interne Details, Request-Bodies werden auf 1 MB begrenzt, unerwünschte CORS-Header fehlen, und ohne gesetzten `API_KEY` ist die API bis auf `/healthz` gesperrt.

Allerdings gibt es ein sicherheitsrelevantes Acceptanz-Kriterium, das im aktuellen Stand verletzt wird: Die Reihenfolge der Middleware-Kette führt dazu, dass ein POST/PUT mit falschem `Content-Type` und zu großem Body mit `413` statt `415` beantwortet wird. Das ist ein fixbarer Konformitätsfehler. Daneben bestehen niedrige bis mittlere Datenschutz- und CRA-Hinweise, vor allem zur Übertragung der User-ID im Query-String und zur notwendigen Dokumentation in den nicht einsehbaren Dateien `README.md`, `SECURITY.md` und `COMPLIANCE.md`.

Die Dateien `COMPLIANCE.md`, `SECURITY.md` und `README.md` sind auf dem Branch vorhanden, aber nicht einsehbar. Die unten genannten Dokumentationsbefunde sind daher als Prüf-/Ergänzungsaufträge zu verstehen, nicht als bestätigtes Fehlen.

---

## 1. DSGVO / Datenschutz

### Befund G-1 — User-ID wird als Query-Parameter übertragen (mittel)

**Betroffene Datei:** `handlers_evaluate.go`  
**Stelle:** `user := r.URL.Query().Get("user")`

Die eigene Logging-Middleware ist sauber und protokolliert nur `r.URL.Path`. Trotzdem ist die Übermittlung der User-ID im Query-String ein datenschutzrechtliches Restrisiko: Vorgelagerte Reverse-Proxys, Loadbalancer oder Access-Logs protokollieren Query-Strings häufig vollständig. Damit könnte die User-ID außerhalb des eigentlichen Dienstes in Logs landen.

**Konkrete Abhilfe:**  
In `handlers_evaluate.go` vorrangig einen Header auslesen, mit Query-String nur als Fallback:

```go
user := r.Header.Get("X-User-ID")
if user == "" {
    user = r.URL.Query().Get("user")
}
```

Zusätzlich in `README.md` dokumentieren, dass Clients `X-User-ID` verwenden sollen und vorgelagerte Systeme keine Query-Logs schreiben dürfen.

---

### Befund G-2 — Rechtsgrundlage der transienten User-ID-Verarbeitung nicht im sichtbaren Code dokumentiert (niedrig)

**Betroffene Datei:** `COMPLIANCE.md` (nicht einsehbar)

Der Service verarbeitet die User-ID ausschließlich transient zur Hash-Berechnung. Er speichert sie nicht und protokolliert sie nicht. Die Rechtsgrundlage dafür muss der Verantwortliche festlegen; sie ist im Code naturgemäß nicht implementierbar.

**Konkrete Abhilfe:**  
In `COMPLIANCE.md` einen Absatz aufnehmen, z. B.:

> „Verarbeitung personenbezogener Daten: Die `user`-ID wird ausschließlich transient zur Berechnung des Rollout-Hashs verwendet. Sie wird nicht gespeichert, nicht persistiert und nicht in Logs geschrieben. Rechtsgrundlage: Art. 6 Abs. 1 lit. b DSGVO (Vertragserfüllung) bzw. Art. 6 Abs. 1 lit. f DSGVO (berechtigtes Interesse an deterministischer Feature-Auslieferung).“

---

### Befund G-3 — Freitextfeld `description` ohne Längenbegrenzung (niedrig)

**Betroffene Dateien:** `handlers_create.go`, `handlers_update.go`

Das Flag-Objekt speichert `Description` als beliebig langes Freitextfeld bis zur 1-MB-Body-Grenze. Da Betreiber dort versehentlich personenbezogene Daten ablegen könnten, empfiehlt sich eine fachliche Längenbegrenzung.

**Konkrete Abhilfe:**  
In `handlers_create.go` und `handlers_update.go` eine Prüfung ergänzen, z. B.:

```go
if len(f.Description) > 512 {
    // 400: description must not exceed 512 characters
}
```

Alternativ in der API-Dokumentation ausdrücklich verbieten, personenbezogene Daten in `description` zu speichern.

---

## 2. EU Cyber Resilience Act (CRA)

### Befund C-1 — AC-13 verletzt: falscher Content-Type bei zu großem Body führt zu 413 statt 415 (hoch)

**Betroffene Dateien:** `main.go`, `auth_test.go`  
**Stelle:**  
```go
handler := Recover(Logging(BodyLimit(ContentType(RequireAuth(mux)))))
```

Die Middleware-Kette führt `BodyLimit` vor `ContentType` aus. Dadurch antwortet ein POST oder PUT mit nicht-JSON-`Content-Type` und einem Body über 1 MB mit `413 Request Entity Too Large`, obwohl AC-13 verlangt, dass andere Content-Types immer mit `415 Unsupported Media Type` beantwortet werden.

Das ist ein expliziter Verstoß gegen ein sicherheitsbezogenes Acceptanz-Kriterium und sollte vor Auslieferung behoben werden.

**Konkrete Abhilfe:**  
In `main.go` und in `auth_test.go` (`newTestHandler`) die Kette auf die fachlich richtige Reihenfolge umstellen:

```go
handler := Logging(Recover(RequireAuth(ContentType(BodyLimit(mux)))))
```

Begründung für die Reihenfolge:

- `Logging` außen: stellt sicher, dass auch bei Panics jeder Request mit Status 500 protokolliert wird.
- `RequireAuth` vor `ContentType`/`BodyLimit`: verhindert, dass unauthentifizierte Clients unnötig Body-Daten verarbeiten.
- `ContentType` vor `BodyLimit`: stellt sicher, dass ein nicht-JSON-`Content-Type` sofort mit 415 abgelehnt wird, ohne vorher den Body vollständig zu lesen.

---

### Befund C-2 — Kein TLS und kein sichtbarer Hinweis zur Transportverschlüsselung (mittel)

**Betroffene Datei:** `main.go`

Der Server bindet an `127.0.0.1:8080` und nutzt `http.ListenAndServe` ohne TLS. Das ist für Loopback-Betrieb sicher, aber sobald die Bind-Adresse auf eine Netzwerkschnittstelle geändert würde, liefen API-Key und User-ID im Klartext.

**Konkrete Abhilfe:**  
In `README.md` oder `SECURITY.md` einen Abschnitt ergänzen:

> „Transportverschlüsselung: Der Dienst bindet absichtlich an `127.0.0.1`. Soll er netzwerkweit erreichbar sein, ist zwingend eine TLS-Terminierung vorzuschalten. Der Dienst darf nicht unverschlüsselt an öffentliche Netze gebunden werden.“

---

### Befund C-3 — Authentifizierung greift erst nach Body-Prüfung (niedrig)

**Betroffene Dateien:** `main.go`, `auth_test.go`

Durch die bestehende Kette `BodyLimit(ContentType(RequireAuth(...)))` kann ein unauthentifizierter Client bei POST/PUT bis zu 1 MB an den Dienst senden, bevor die Authentifizierung greift.

**Konkrete Abhilfe:**  
Wird durch die unter C-1 genannte Umstellung auf `RequireAuth(ContentType(BodyLimit(...))` behoben.

---

### Befund C-4 — Update-, Support- und Sicherheitsmeldeprozess nicht im sichtbaren Code erkennbar (niedrig)

**Betroffene Dateien:** `README.md`, `SECURITY.md` (nicht einsehbar)

Ein Go-Backend braucht keinen eingebauten Update-Mechanismus; Updates erfolgen durch erneutes Deployment. Für CRA-Konformität muss der Hersteller aber dokumentieren, wie Sicherheitsupdates bereitgestellt werden und wie Sicherheitslücken gemeldet werden können.

**Konkrete Abhilfe:**  
In `SECURITY.md` sicherstellen, dass folgende Abschnitte enthalten sind:

- „Supported versions“
- „Security updates“
- „Reporting vulnerabilities“

Falls diese Abschnitte bereits vorhanden sind, ist kein Handlungsbedarf gegeben.

---

## 3. EU AI Act

Nicht anwendbar. Der Service enthält keine KI-Funktion, kein maschinelles Lernen und keine automatisierte Einzelfallentscheidung im Sinne des AI Act. Die Feature-Flag-Entscheidung ist deterministisch und regelbasiert.

---

## 4. Pflichttexte & UI

Nicht anwendbar. Das Produkt ist ein reines Backend ohne öffentliche Web-UI. Es bestehen keine Pflichten zu Impressum, Cookie-Banner, Consent-Management oder verbraucherschutzrechtlichen Widerrufsbelehrungen im Code.

---

## 5. Barrierefreiheit

Nicht anwendbar. Das Produkt stellt keine öffentliche Web-Oberfläche bereit. WCAG/BITV/EAA-Pflichten greifen für dieses reine REST-Backend nicht.

---

## Fazit

Der Code ist in den zentralen Datenschutz- und Sicherheitsanforderungen solide: keine PII in Logs, keine Stacktraces, sichere Defaults bei fehlendem API-Key, Body-Limit, Content-Type-Prüfung und keine CORS-Header. Der einzige echte Konformitätsfehler ist die Middleware-Reihenfolge in `main.go`/`auth_test.go`, die AC-13 verletzt und zugleich zwei kleinere Härtungsziele beeinträchtigt. Nach Umsetzung dieser Umstellung und Ergänzung der unter G-2, C-2 und C-4 genannten Dokumentationspunkte ist der Stand marktreif.