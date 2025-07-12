# Klassenmischer: ein Schüler-Gruppierungstool 

Dieses Go-Programm hilft Ihnen, Schüler basierend auf einer Konfigurationsdatei in Gruppen einzuteilen.  
Es berücksichtigt dabei spezielle Einschränkungen, um sicherzustellen, dass bestimmte Schüler nicht in derselben Gruppe landen.  
Das Tool bietet verschiedene Gruppierungs-Szenarien und versucht, so viele Schüler wie möglich gemäß den Regeln zu gruppieren.

## Funktionen

* **Flexible Gruppierung:** Unterstützt die Bildung von 2er-, 3er- und 4er-Gruppen.
* **Anpassung bei Restschülern:** Intelligente Anpassung (z.B. Bildung von 3er-Gruppen im 2er-Szenario oder 2er/4er/5er-Gruppen in den anderen Szenarien), um möglichst wenige Schüler ungruppiert zu lassen.
* **Konfliktmanagement:** Berücksichtigt definierte Einschränkungen, wer nicht mit wem in eine Gruppe soll.
* **Einschränkungs-Validierung:** Prüft die Konfigurationsdatei auf symmetrische Einschränkungen, um Logikfehler zu vermeiden.
* **Einfache Konfiguration:** Alle Schülerlisten und Einschränkungen werden über eine `klasse.toml`-Datei verwaltet.
* **Plattformübergreifend:** Läuft auf Windows, macOS und Linux.



## Installation und Ausführung

Laden Sie das passende Programm aus dem Ordner `programme`.  
Starten Sie es mit Doppelklick oder im Terminal.  
Eventuell müssen Sie die Datei mit `chmod +x`ausführbar gemacht werden.


## Konfiguration (`klasse.toml`)

Beim ersten Start des Programms wird automatisch eine `klasse.toml`-Datei im selben Verzeichnis erstellt.  
Diese Datei enthält eine Beispiel-Schülerliste und ein Beispiel für Konflikte.

**Wichtig:** Bitte passen Sie diese Datei an Ihre tatsächlichen Schülerlisten und gewünschten Einschränkungen an.

### Beispiel `klasse.toml`:

```toml
schuelerliste = ["Alice", "Bob", "Charlie", "David", "Eve", "Frank"]

# Hier können Sie Einschränkungen definieren, wer nicht mit wem in eine Gruppe soll.
# Beispiel: "Schueler A" = ["Schueler B", "Schueler C"]
# Achten Sie auf symmetrische Einschränkungen! Wenn "X" nicht mit "Y" soll, muss auch "Y" nicht mit "X" wollen.
"Alice" = ["Bob"]
"Bob" = ["Alice"]

# Bitte passen Sie die 'schuelerliste' und 'Konflikte' oben an Ihre Bedürfnisse an.
```

**Hinweise zu den Einschränkungen:**

* Jede Einschränkung sollte symmetrisch sein.  
* Wenn "Max" nicht mit "Lisa" in eine Gruppe soll, müssen Sie sowohl `"Max" = ["Lisa"]` als auch `"Lisa" = ["Max"]` definieren. Das Programm prüft dies und gibt eine Warnung aus, falls Asymmetrien gefunden werden.
* Schüler, die keine Einschränkungen haben, müssen nicht in der `klasse.toml` aufgeführt werden.

## Funktionsweise

Das Programm durchläuft folgende Schritte:

1. **Konfiguration laden:** Versucht, `klasse.toml` zu finden und zu lesen. Wenn die Datei nicht existiert, wird eine neue Musterdatei erstellt und das Programm beendet sich mit einem Hinweis.
2. **Einschränkungen-Prüfung:** Überprüft die definierten Einschränkungen auf Symmetrie und gibt eine Warnung aus, wenn Inkonsistenzen gefunden werden.
3. **Gruppenbildung:** Versucht in drei verschiedenen Szenarien (2er-, 3er- und 4er-Gruppen) die bestmögliche Gruppierung zu finden. Jedes Szenario wird mehrfach (standardmäßig 1000 Mal) mit zufällig gemischten Schülerlisten wiederholt, um optimale Ergebnisse zu erzielen.
4. **Ergebnisse anzeigen:** Die gebildeten Gruppen und eventuell übrig gebliebene ungruppierte Schüler werden auf der Konsole ausgegeben.


## Lizenz

Dieses Programm wird unter der MIT-Lizenz veröffentlicht. Details finden Sie in der `LICENSE`-Datei (falls vorhanden) oder Sie können sie hier hinzufügen.

