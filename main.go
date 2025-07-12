package main // Deklariert das Paket als 'main', was bedeutet, dass dies ein ausführbares Programm ist.

// ############################################################################################
import ( // Importiert notwendige Pakete.
	"fmt"          // Für formatierte Ein- und Ausgabe (z.B. Drucken auf die Konsole, `fmt.Scanln`).
	"log"          // Für Logging-Ausgaben, besonders nützlich für Debugging-Informationen.
	"math/rand"    // Für Zufallszahlen-Operationen, hier zum Mischen von Schülerlisten.
	"os"           // Bietet Schnittstellen zum Betriebssystem (z.B. Dateisystem-Operationen, Beenden des Programms).
	"path/filepath" // Für plattformunabhängige Pfadmanipulation (z.B. Join, Dir).
	"sort"         // Zum Sortieren von Slices, hier für das Prüfen symmetrischer Constraints.
	"strings"      // Für String-Manipulationen (z.B. Join, Contains, HasPrefix).

	"github.com/pelletier/go-toml" // Externe Bibliothek zum Lesen und Schreiben von TOML-Dateien.
)

// ############################################################################################
// ConfigFileCreatedError ist ein benutzerdefinierter Fehlertyp.
// Er wird zurückgegeben, wenn die Konfigurationsdatei ('klasse.toml') nicht gefunden wurde
// 		und eine neue Musterdatei erstellt werden musste.
// Dies ist kein "echter" Fehler im Sinne eines Programmabbruchs, 
// 		sondern eine Information an den aufrufenden Code, dass eine Benutzeraktion erforderlich ist.
type ConfigFileCreatedError struct {
	FilePath string // Speichert den Pfad der neu erstellten Datei.
}

// Error implementiert die 'error'-Schnittstelle für ConfigFileCreatedError.
// Dies ermöglicht es, eine Instanz dieses Typs als Fehler zurückzugeben.
func (e *ConfigFileCreatedError) Error() string {
	// return fmt.Sprintf("\nBitte passen Sie diese Datei an und starten Sie das Programm neut. (Pfad: %s)", e.FilePath)
	return fmt.Sprintln("\nBitte passe diese Datei an und starte das Programm neu.")
}

// ############################################################################################
// Config repräsentiert die Struktur der 'klasse.toml'-Konfigurationsdatei.
// Die Tags `toml:"..."` weisen die go-toml-Bibliothek an, welche Felder in der TOML-Datei
// 		zu welchen Feldern in dieser Go-Struktur gehören.
type Config struct {
	Schuelerliste []string `toml:"schuelerliste"` 
	// Ein Slice von Strings für die Namen der Schüler.
	Constraints   map[string][]string `toml:"-"`  
	// Eine Map für Einschränkungen. Der Tag `toml:"-"` bedeutet,
	// 		dass dieses Feld von der TOML-Bibliothek ignoriert werden soll.
	// Constraints werden manuell aus der TOML-Datei geparst.
}

// ############################################################################################
// readTomlConfig versucht, die Konfigurationsdatei zu finden, zu lesen und zu parsen.
// Wenn die Datei nicht existiert, wird eine Musterdatei erstellt 
// 		und ein spezieller Fehler zurückgegeben.
func readTomlConfig(filename string) (*Config, error) {
	var finalConfigPath string // Speichert den endgültigen Pfad, der zum Lesen oder Schreiben verwendet wird.
	fileFound := false         // Flag, das anzeigt, ob eine Konfigurationsdatei gefunden wurde.

	// Variablen für das aktuelle Arbeitsverzeichnis (CWD) und den Pfad der ausführbaren Datei.
	// Diese werden hier deklariert, damit sie im gesamten Funktionsumfang zugänglich sind.
	var currentWorkingDir string
	var getwdErr error
	var execPath string
	var execErr error

	// --- DEBUGGING: Konfigurationsdatei-Pfadsuche ---
	// Dieser Block diente nur zur Ausgabe von Debugging-Informationen über die Pfade und ist nun deaktiviert.
	// log.Println("--- DEBUGGING: Konfigurationsdatei-Pfadsuche ---")

	currentWorkingDir, getwdErr = os.Getwd() // Holt das aktuelle Arbeitsverzeichnis.
	// if getwdErr == nil {
	// 	log.Printf("Aktuelles Arbeitsverzeichnis (CWD): %s", currentWorkingDir)
	// } else {
	// 	log.Printf("Fehler beim Ermitteln des CWD: %v", getwdErr)
	// }

	execPath, execErr = os.Executable() // Holt den vollständigen Pfad zur ausführbaren Datei.
	// if execErr == nil {
	// 	log.Printf("Pfad der ausführbaren Datei: %s", execPath)
	// 	log.Printf("Verzeichnis der ausführbaren Datei: %s", filepath.Dir(execPath)) // Nur das Verzeichnis.
	// } else {
	// 	log.Printf("Fehler beim Ermitteln des Executable-Pfades: %v", execErr)
	// }
	// log.Println("--- ENDE DEBUGGING ---")

	// --- Heuristik für 'go run' Erkennung ---
	// Diese Logik versucht zu erraten, ob das Programm über 'go run' gestartet wurde.
	// 'go run' erstellt oft temporäre Executables an speziellen Orten.
	isGoRunBuild := false
	if execErr == nil { // Nur prüfen, wenn der Executable-Pfad erfolgreich ermittelt wurde.
		execDir := filepath.Dir(execPath)  // Das Verzeichnis, in dem das Executable liegt.
		tempDir := os.TempDir()            // Das systemweite temporäre Verzeichnis.

		// Überprüft auf typische Muster von temporären Go-Builds:
		// 1. Liegt das Executable-Verzeichnis im systemweiten Temp-Verzeichnis?
		// 2. Enthält der Basisname des Executable-Verzeichnisses "go-build" (typisch für Go-Temp-Builds)?
		// 3. Enthält der Pfad "/var/folders/" (typisch für temporäre macOS-Pfade)?
		if strings.HasPrefix(execDir, tempDir) ||
			strings.Contains(filepath.Base(execDir), "go-build") ||
			strings.Contains(execDir, "/var/folders/") {
			isGoRunBuild = true
			// log.Printf("Heuristik: Programm läuft wahrscheinlich über 'go run' (Executable-Pfad: %s)", execPath) // DEBUG
		}
	}
	// --- ENDE Heuristik ---

	// --- Prioritäten beim SUCHEN (Lesen) der Konfigurationsdatei ---

	// 1. Priorität (Lesen): Versuche die Datei im aktuellen Arbeitsverzeichnis (CWD) zu finden.
	// Das ist der Standardort für Entwicklung mit 'go run' oder wenn der Benutzer die Datei hier ablegt.
	cwd := currentWorkingDir // Verwende die bereits ermittelten Werte.
	err := getwdErr
	if err != nil { // Wenn CWD nicht ermittelt werden konnte, ist dieser Pfad unbrauchbar.
		// log.Printf("Warnung: Konnte das aktuelle Arbeitsverzeichnis nicht ermitteln: %v. Versuche, die Konfigurationsdatei im Verzeichnis der ausführbaren Datei zu finden.", err)
	} else {
		pathInCWD := filepath.Join(cwd, filename) // Kombiniert CWD und Dateinamen zu einem vollständigen Pfad.
		// log.Printf("Prüfe Konfigurationsdatei im CWD: %s", pathInCWD) // DEBUG-Ausgabe.
		_, errCWD := os.Stat(pathInCWD)                              // Versucht, Datei-Informationen zu holen (prüft Existenz).
		if errCWD == nil {                                           // Wenn kein Fehler, existiert die Datei.
			finalConfigPath = pathInCWD
			fileFound = true // Datei gefunden!
			// log.Printf("Konfigurationsdatei im CWD gefunden: %s", finalConfigPath) // DEBUG
		} else if os.IsNotExist(errCWD) { // Spezifischer Fehler: Datei existiert nicht.
			// log.Printf("Konfigurationsdatei NICHT im CWD gefunden: %s", pathInCWD) // DEBUG
		} else { // Anderer Fehler (z.B. Berechtigungen) beim Zugriff auf Datei im CWD.
			return nil, fmt.Errorf("❌ Fehler beim Zugriff auf Konfigurationsdatei im aktuellen Verzeichnis '%s': %w", pathInCWD, errCWD)
		}
	}

	// 2. Zweite Priorität (Lesen): Verzeichnis der ausführbaren Datei.
	// Dieser Pfad ist wichtig für kompilierte Programme, die von einem anderen Ort gestartet werden.
	if !fileFound { // Nur prüfen, wenn die Datei im CWD noch nicht gefunden wurde.
		if execErr != nil { // Wenn der Executable-Pfad selbst nicht ermittelt werden konnte.
			// log.Printf("Warnung: Konnte den Pfad der ausführbaren Datei nicht ermitteln: %v", execErr)
		} else {
			// execPath ist jetzt zugänglich, da es im Funktionsumfang deklariert wurde.
			pathFromExe := filepath.Join(filepath.Dir(execPath), filename) // Pfad zum Verzeichnis des Executables.
			// log.Printf("Prüfe Konfigurationsdatei im Executable-Verzeichnis: %s", pathFromExe) // DEBUG
			_, statErr := os.Stat(pathFromExe)                                                 // Prüft Existenz.
			if statErr == nil {                                                                // Wenn kein Fehler, existiert die Datei.
				finalConfigPath = pathFromExe
				fileFound = true // Datei gefunden!
				// log.Printf("Konfigurationsdatei im Executable-Verzeichnis gefunden: %s", finalConfigPath) // DEBUG
			} else if os.IsNotExist(statErr) { // Datei existiert nicht an diesem Ort.
				// log.Printf("Konfigurationsdatei NICHT im Executable-Verzeichnis gefunden: %s", pathFromExe) // DEBUG
			} else { // Anderer Fehler beim Zugriff.
				return nil, fmt.Errorf("❌ Fehler beim Zugriff auf Konfigurationsdatei am Executable-Pfad '%s': %w", pathFromExe, statErr)
			}
		}
	}

	// --- Lesen und Parsen der gefundenen Datei ---
	if fileFound { // Wenn eine Datei an einem der oben genannten Orte gefunden wurde.
		// log.Printf("Lese Konfigurationsdatei von: %s", finalConfigPath) // DEBUG
		data, err := os.ReadFile(finalConfigPath)                        // Liest den gesamten Dateiinhalt.
		if err != nil {
			return nil, fmt.Errorf("❌ Fehler beim Lesen der Datei %s: %w", finalConfigPath, err)
		}

		tree, err := toml.LoadBytes(data) // Parsen der TOML-Daten in eine Baumstruktur.
		if err != nil {
			return nil, fmt.Errorf("❌ Fehler beim Parsen der TOML-Datei: %w", err)
		}

		var config Config // Erstellt eine leere Config-Struktur.
		// Entpackt die einfachen Felder (wie 'schuelerliste') direkt aus dem TOML-Baum in die Struktur.
		if err := tree.Unmarshal(&config); err != nil {
			return nil, fmt.Errorf("❌ Fehler beim Entpacken der Schülerliste aus der TOML-Datei: %w", err)
		}

		// Constraints müssen manuell aus dem TOML-Baum extrahiert werden,
		// da sie dynamische Schlüssel haben und nicht direkt mit 'toml:"-"' gemarshallt werden.
		config.Constraints = make(map[string][]string) // Initialisiert die Constraints-Map.
		for _, key := range tree.Keys() {             // Iteriert über alle Schlüssel in der TOML-Datei.
			if key != "schuelerliste" { // Der Schlüssel "schuelerliste" wurde bereits behandelt.
				val := tree.Get(key) // Holt den Wert für den aktuellen Schlüssel.
				// Überprüft, ob der Wert ein Slice von Interfaces ist (was einem TOML-Array entspricht).
				if valSlice, ok := val.([]interface{}); ok {
					var forbiddenStudents []string
					for _, item := range valSlice { // Iteriert über die Elemente des Arrays.
						if studentName, isString := item.(string); isString { // Prüft, ob das Element ein String ist.
							forbiddenStudents = append(forbiddenStudents, studentName) // Fügt den Namen hinzu.
						} else {
							log.Printf("❗️ Warnung: Konflikte für '%s' enthält einen Nicht-String-Wert: %v", key, item)
						}
					}
					config.Constraints[key] = forbiddenStudents // Fügt die Einschränkung zur Map hinzu.
				} else {
					log.Printf("❗️ Warnung: Unerwarteter Typ für Schlüssel '%s' in TOML-Datei. Erwartet wurde ein Array, gefunden: %T", key, val)
				}
			}
		}
		return &config, nil // Gibt die befüllte Konfiguration zurück.
	}

	// --- Erstellen der Musterdatei, wenn keine Datei gefunden wurde ---
	// Die Priorität beim Erstellen hängt davon ab, wie das Programm gestartet wurde.
	var pathToCreate string // Der Pfad, an dem die Musterdatei erstellt werden soll.

	// 1. Priorität (Erstellen): Wenn es 'go run' ist UND CWD ermittelbar.
	// In diesem Fall ist das CWD der logischste Ort für die Erstellung der Datei.
	if isGoRunBuild && getwdErr == nil {
		pathToCreate = filepath.Join(currentWorkingDir, filename)
		// log.Printf("Logik: 'go run' erkannt, erstelle Musterdatei im aktuellen Arbeitsverzeichnis (CWD): '%s'", pathToCreate)
	} else if execErr == nil { // 2. Priorität (Erstellen): Ansonsten, wenn Executable-Pfad bekannt (typisch für kompilierte Apps).
		// Dies deckt den Fall ab, dass das Programm kompiliert wurde und sein CWD woanders liegt (z.B. Home-Verzeichnis).
		pathToCreate = filepath.Join(filepath.Dir(execPath), filename)
		// log.Printf("Logik: Kein 'go run' oder CWD nicht nutzbar, erstelle Musterdatei im Executable-Verzeichnis: '%s'", pathToCreate)
	} else if getwdErr == nil { // 3. Priorität (Erstellen): Fallback, falls Executable-Pfad unbekannt, aber CWD bekannt.
		// Dies ist ein seltener Fall, aber eine gute Absicherung.
		pathToCreate = filepath.Join(currentWorkingDir, filename)
		// log.Printf("Logik: Executable-Pfad nicht ermittelbar, aber CWD bekannt. Erstelle im CWD: '%s'", pathToCreate)
	} else { // Letzter Ausweg: Weder CWD noch Executable-Pfad ermittelbar.
		// Wenn gar keine Pfadinformationen vorhanden sind, versuchen wir es mit dem reinen Dateinamen.
		pathToCreate = filename
		// log.Printf("Logik: Weder CWD noch Executable-Pfad ermittelbar. Versuche, Musterdatei als '%s' zu erstellen (Potenziell unsicher).", pathToCreate)
	}
	finalConfigPath = pathToCreate // Setzt den ermittelten Pfad für die Erstellung.

	fmt.Printf("❗️ '%s' nicht gefunden. \nErstelle eine Musterdatei.", finalConfigPath)
	// log.Printf("Erstelle Muster-Konfigurationsdatei unter: %s", finalConfigPath) // DEBUG

	// Standard-Schülerliste und Beispiel-Constraints für die Musterdatei.
	defaultSchuelerliste := []string{
		"Schueler 1", "Schueler 2", "Schueler 3", "Schueler 4", "Schueler 5",
		"Schueler 6", "Schueler 7", "Schueler 8", "Schueler 9", "Schueler 10",
	}
	sampleConstraints := map[string][]string{
		"Schueler 1": {"Schueler 2", "Schueler 3"},
		"Schueler 2": {"Schueler 1"}, // Symmetrische Einschränkung als Beispiel.
		"Schueler 3": {"Schueler 1"},
	}

	// Erstellt den Inhalt der Musterdatei als String.
	var sb strings.Builder                                                                    // Effizienter String-Builder.
	sb.WriteString(fmt.Sprintf("schuelerliste = %s\n\n", formatStringSliceToTomlArray(defaultSchuelerliste))) // Schülerliste als TOML-Array.
	sb.WriteString("# Hier kannst du Einschränkungen definieren, wer nicht mit wem in eine Gruppe soll.\n")
	sb.WriteString("# Beispiel: \"Schueler A\" = [\"Schueler B\", \"Schueler C\"]\n")
	sb.WriteString("# Achte auf symmetrische Einschränkungen! Wenn \"X\" nicht mit \"Y\" soll, muss auch \"Y\" nicht mit \"X\" wollen.\n")
	for student, forbidden := range sampleConstraints { // Fügt die Beispiel-Constraints hinzu.
		sb.WriteString(fmt.Sprintf("%q = %s\n", student, formatStringSliceToTomlArray(forbidden)))
	}
	sb.WriteString("\n# Bitte passe die 'schuelerliste' und 'Konflikte' oben an deine Bedürfnisse an.\n")

	// Schreibt den erstellten Inhalt in die Datei.
	err = os.WriteFile(finalConfigPath, []byte(sb.String()), 0644) // 0644 sind Dateiberechtigungen (Lesen/Schreiben für Besitzer, nur Lesen für andere).
	if err != nil {
		return nil, fmt.Errorf("❌ Fehler beim Schreiben der Muster-Konfigurationsdatei '%s': %w", finalConfigPath, err)
	}
	// Gibt den speziellen Fehler zurück, um anzuzeigen, dass eine Datei erstellt wurde.
	return nil, &ConfigFileCreatedError{FilePath: finalConfigPath}
}

// ############################################################################################
// formatStringSliceToTomlArray ist eine Hilfsfunktion, die einen Slice von Strings
// in ein TOML-konformes Array-String-Format umwandelt (z.B. ["A", "B", "C"]).
func formatStringSliceToTomlArray(slice []string) string {
	quotedStrings := make([]string, len(slice))
	for i, s := range slice {
		quotedStrings[i] = fmt.Sprintf("%q", s) // Fügt Anführungszeichen um jeden String.
	}
	return "[" + strings.Join(quotedStrings, ", ") + "]" // Verbindet sie mit Komma und Leerzeichen.
}

// ############################################################################################
// isValidGroup überprüft, ob eine gegebene Gruppe von Schülern gültig ist,
// basierend auf den definierten Einschränkungen (Constraints).
// Eine Gruppe ist ungültig, wenn Schüler in ihr sind, die nicht zusammenarbeiten dürfen.
func isValidGroup(group []string, constraints map[string][]string) bool {
	// Iteriert über jedes mögliche Paar von Schülern innerhalb der Gruppe.
	for i, studentA := range group {
		forbiddenList, exists := constraints[studentA] // Holt die Liste der Schüler, mit denen studentA nicht zusammenarbeiten darf.
		if !exists {
			continue // Wenn studentA keine Einschränkungen hat, überspringe ihn.
		}

		for j, studentB := range group {
			if i == j {
				continue // Überspringe den Vergleich eines Schülers mit sich selbst.
			}

			// Prüft, ob studentB in der Verbotsliste von studentA ist.
			for _, forbiddenStudent := range forbiddenList {
				if studentB == forbiddenStudent {
					return false // Ungültige Gruppe gefunden!
				}
			}
			// Zusätzlich prüft man die umgekehrte Richtung für Symmetrie (redundant, wenn Symmetrie geprüft wurde, aber sicherheitshalber).
			forbiddenListB, existsB := constraints[studentB]
			if existsB {
				for _, forbiddenStudent := range forbiddenListB {
					if studentA == forbiddenStudent {
						return false // Ungültige Gruppe gefunden!
					}
				}
			}
		}
	}
	return true // Wenn keine Konflikte gefunden wurden, ist die Gruppe gültig.
}

// ############################################################################################
// attemptToFormGroupsOfSize versucht, so viele Gruppen einer bestimmten Zielgröße wie möglich zu bilden.
// Es wählt zufällig Schüler aus und prüft, ob die Gruppe gültig ist.
func attemptToFormGroupsOfSize(targetSize int, studentsPool []string, usedStudents map[string]bool,
	existingGroups [][]string, constraints map[string][]string) ([][]string, map[string]bool) {

	// Erstellt Kopien der aktuellen Gruppen und verwendeten Schüler, um Änderungen rückgängig machen zu können,
	// falls eine Iteration nicht zu besseren Ergebnissen führt.
	currentGroups := make([][]string, len(existingGroups))
	copy(currentGroups, existingGroups)
	currentUsedStudents := make(map[string]bool)
	for k, v := range usedStudents {
		currentUsedStudents[k] = v
	}

	for { // Endlosschleife, die abbricht, wenn keine weiteren Gruppen gebildet werden können.
		availableStudents := []string{}
		for _, s := range studentsPool {
			if !currentUsedStudents[s] { // Sammelt alle noch nicht verwendeten Schüler.
				availableStudents = append(availableStudents, s)
			}
		}

		if len(availableStudents) < targetSize { // Wenn nicht genug Schüler für eine weitere Gruppe übrig sind.
			break
		}

		// Mischt die Liste der verfügbaren Schüler, um zufällige Gruppen zu bilden.
		rand.Shuffle(len(availableStudents), func(i, j int) {
			availableStudents[i], availableStudents[j] = availableStudents[j], availableStudents[i]
		})

		foundGroupInThisIteration := false
		if len(availableStudents) >= targetSize {
			// Wählt die ersten 'targetSize' Schüler für eine potenzielle Gruppe.
			potentialGroup := make([]string, targetSize)
			copy(potentialGroup, availableStudents[:targetSize])

			if isValidGroup(potentialGroup, constraints) { // Prüft, ob die Gruppe gültig ist.
				currentGroups = append(currentGroups, potentialGroup) // Fügt die Gruppe hinzu.
				for _, s := range potentialGroup {
					currentUsedStudents[s] = true // Markiert die Schüler als verwendet.
				}
				foundGroupInThisIteration = true
			}
		}

		if !foundGroupInThisIteration { // Wenn in dieser Runde keine Gruppe gebildet werden konnte, ist Schluss.
			break
		}
	}
	return currentGroups, currentUsedStudents // Gibt die gebildeten Gruppen und die verwendeten Schüler zurück.
}

// ############################################################################################
// tryIntegrateIntoExistingGroup versucht, einen einzelnen "einsamen" Schüler
// in eine bestehende Gruppe zu integrieren, um eine neue Zielgröße zu erreichen.
func tryIntegrateIntoExistingGroup(lonelyStudent string, targetGroupSize int, newGroupSize int,
	existingGroups [][]string, usedStudents map[string]bool, constraints map[string][]string) (bool, [][]string, map[string]bool) {

	// Erstellt Kopien der Daten, um keine unerwünschten Seiteneffekte zu verursachen.
	groupsCopy := make([][]string, len(existingGroups))
	copy(groupsCopy, existingGroups)
	usedStudentsCopy := make(map[string]bool)
	for k, v := range usedStudents {
		usedStudentsCopy[k] = v
	}

	// Mischt die Reihenfolge der Gruppen, um Zufälligkeit bei der Integration zu gewährleisten.
	indices := make([]int, len(groupsCopy))
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	for _, idx := range indices { // Iteriert über die Gruppen.
		group := groupsCopy[idx]
		if len(group) == targetGroupSize { // Findet eine Gruppe der ursprünglichen Zielgröße (z.B. 2er-Gruppe bei 2er-Szenario).
			potentialNewGroup := append([]string{}, group...) // Kopiert die Gruppe.
			potentialNewGroup = append(potentialNewGroup, lonelyStudent) // Fügt den einsamen Schüler hinzu.

			if isValidGroup(potentialNewGroup, constraints) { // Prüft, ob die neue, größere Gruppe gültig ist.
				groupsCopy[idx] = potentialNewGroup        // Aktualisiert die Gruppe.
				usedStudentsCopy[lonelyStudent] = true     // Markiert den Schüler als verwendet.
				return true, groupsCopy, usedStudentsCopy // Erfolgreich integriert!
			}
		}
	}
	return false, existingGroups, usedStudents // Konnte nicht integrieren.
}

// ############################################################################################
// formTwoPersonGroups versucht, primär 2er-Gruppen zu bilden.
// Es hat spezielle Logik, um einen einzelnen Restschüler in eine 3er-Gruppe zu verwandeln.
func formTwoPersonGroups(allStudents []string, constraints map[string][]string) ([][]string, []string) {
	studentsToGroup := make([]string, len(allStudents))
	copy(studentsToGroup, allStudents)
	rand.Shuffle(len(studentsToGroup), func(i, j int) { // Mischt die Schülerliste.
		studentsToGroup[i], studentsToGroup[j] = studentsToGroup[j], studentsToGroup[i] // Korrekter Tausch.
	})

	var groups [][]string        // Die Liste der gebildeten Gruppen.
	usedStudents := make(map[string]bool) // Map, um zu verfolgen, welche Schüler verwendet wurden.

	// Versucht, so viele 2er-Gruppen wie möglich zu bilden.
	groups, usedStudents = attemptToFormGroupsOfSize(2, studentsToGroup, usedStudents, groups, constraints)

	var currentlyUngrouped []string // Schüler, die nach der Hauptbildung übrig sind.
	for _, student := range studentsToGroup {
		if !usedStudents[student] {
			currentlyUngrouped = append(currentlyUngrouped, student)
		}
	}

	// Spezialfall: Ein einzelner ungepaarter Schüler übrig.
	if len(currentlyUngrouped) == 1 {
		lonelyStudent := currentlyUngrouped[0]
		// Versucht, den einzelnen Schüler in eine 2er-Gruppe zu integrieren, um eine 3er-Gruppe zu bilden.
		integrated, updatedGroups, updatedUsedStudents := tryIntegrateIntoExistingGroup(lonelyStudent, 2, 3, groups, usedStudents, constraints)
		if integrated {
			groups = updatedGroups
			usedStudents = updatedUsedStudents
			currentlyUngrouped = []string{} // Der Schüler ist jetzt nicht mehr ungepaart.
		}
	}

	var finalUngrouped []string // Endgültige Liste der ungepaarten Schüler.
	for _, student := range allStudents {
		if !usedStudents[student] {
			finalUngrouped = append(finalUngrouped, student)
		}
	}
	return groups, finalUngrouped
}

// ############################################################################################
// formThreePersonGroups versucht, primär 3er-Gruppen zu bilden.
// Es hat spezielle Logik für 1 oder 2 Restschüler.
func formThreePersonGroups(allStudents []string, constraints map[string][]string) ([][]string, []string) {
	studentsToGroup := make([]string, len(allStudents))
	copy(studentsToGroup, allStudents)
	rand.Shuffle(len(studentsToGroup), func(i, j int) { // Mischt die Schülerliste.
		studentsToGroup[i], studentsToGroup[j] = studentsToGroup[j], studentsToGroup[i] // Korrekter Tausch.
	})

	var groups [][]string
	usedStudents := make(map[string]bool)

	// Versucht, so viele 3er-Gruppen wie möglich zu bilden.
	groups, usedStudents = attemptToFormGroupsOfSize(3, studentsToGroup, usedStudents, groups, constraints)

	var currentlyUngrouped []string
	for _, student := range studentsToGroup {
		if !usedStudents[student] {
			currentlyUngrouped = append(currentlyUngrouped, student)
		}
	}

	// Spezialfall: Ein einzelner ungepaarter Schüler.
	if len(currentlyUngrouped) == 1 {
		lonelyStudent := currentlyUngrouped[0]
		// Versucht, ihn in eine 3er-Gruppe zu integrieren, um eine 4er-Gruppe zu bilden.
		integrated, updatedGroups, updatedUsedStudents := tryIntegrateIntoExistingGroup(lonelyStudent, 3, 4, groups, usedStudents, constraints)
		if integrated {
			groups = updatedGroups
			usedStudents = updatedUsedStudents
			currentlyUngrouped = []string{}
		}
	} else if len(currentlyUngrouped) == 2 { // Spezialfall: Zwei ungepaarte Schüler.
		potentialGroup := currentlyUngrouped // Bilden eine 2er-Gruppe aus den Restschülern.
		if isValidGroup(potentialGroup, constraints) {
			groups = append(groups, potentialGroup) // Fügt die 2er-Restgruppe hinzu.
			for _, s := range potentialGroup {
				usedStudents[s] = true
			}
			currentlyUngrouped = []string{}
		}
	}

	var finalUngrouped []string
	for _, student := range allStudents {
		if !usedStudents[student] {
			finalUngrouped = append(finalUngrouped, student)
		}
	}
	return groups, finalUngrouped
}

// ############################################################################################
// formFourPersonGroups versucht, primär 4er-Gruppen zu bilden.
// Es hat spezielle Logik für 1, 2 oder 3 Restschüler.
func formFourPersonGroups(allStudents []string, constraints map[string][]string) ([][]string, []string) {
	studentsToGroup := make([]string, len(allStudents))
	copy(studentsToGroup, allStudents)
	rand.Shuffle(len(studentsToGroup), func(i, j int) { // Mischt die Schülerliste.
		studentsToGroup[i], studentsToGroup[j] = studentsToGroup[j], studentsToGroup[i] // Korrekter Tausch.
	})

	var groups [][]string
	usedStudents := make(map[string]bool)

	// Versucht, so viele 4er-Gruppen wie möglich zu bilden.
	groups, usedStudents = attemptToFormGroupsOfSize(4, studentsToGroup, usedStudents, groups, constraints)

	var currentlyUngrouped []string
	for _, student := range studentsToGroup {
		if !usedStudents[student] {
			currentlyUngrouped = append(currentlyUngrouped, student)
		}
	}

	// Spezialfall: Ein einzelner ungepaarter Schüler.
	if len(currentlyUngrouped) == 1 {
		lonelyStudent := currentlyUngrouped[0]
		// Versucht, ihn in eine 4er-Gruppe zu integrieren, um eine 5er-Gruppe zu bilden.
		integrated, updatedGroups, updatedUsedStudents := tryIntegrateIntoExistingGroup(lonelyStudent, 4, 5, groups, usedStudents, constraints)
		if integrated {
			groups = updatedGroups
			usedStudents = updatedUsedStudents
			currentlyUngrouped = []string{}
		}
	} else if len(currentlyUngrouped) == 2 { // Spezialfall: Zwei ungepaarte Schüler.
		// Versucht, beide einzeln in bestehende 4er-Gruppen zu integrieren, um 5er-Gruppen zu bilden.
		tempGroups := make([][]string, len(groups))
		copy(tempGroups, groups)
		tempUsedStudents := make(map[string]bool)
		for k, v := range usedStudents {
			tempUsedStudents[k] = v
		}

		integratedCount := 0
		var remainingUngrouped []string // Schüler, die auch nach diesen Versuchen ungepaart bleiben.

		lonelyStudent1 := currentlyUngrouped[0]
		integrated1, updatedGroups1, updatedUsedStudents1 := tryIntegrateIntoExistingGroup(lonelyStudent1, 4, 5, tempGroups, tempUsedStudents, constraints)

		if integrated1 {
			groups = updatedGroups1
			usedStudents = updatedUsedStudents1
			integratedCount++

			if len(currentlyUngrouped) > 1 {
				lonelyStudent2 := currentlyUngrouped[1]
				integrated2, updatedGroups2, updatedUsedStudents2 := tryIntegrateIntoExistingGroup(lonelyStudent2, 4, 5, groups, usedStudents, constraints)
				if integrated2 {
					groups = updatedGroups2
					usedStudents = updatedUsedStudents2
					integratedCount++
				} else {
					remainingUngrouped = append(remainingUngrouped, lonelyStudent2)
				}
			}
		} else {
			// Wenn der erste Schüler nicht integriert werden konnte, bleiben beide ungepaart.
			remainingUngrouped = append(remainingUngrouped, currentlyUngrouped...)
		}

		currentlyUngrouped = remainingUngrouped

	} else if len(currentlyUngrouped) == 3 { // Spezialfall: Drei ungepaarte Schüler.
		potentialGroup := currentlyUngrouped // Bilden eine 3er-Gruppe aus den Restschülern.
		if isValidGroup(potentialGroup, constraints) {
			groups = append(groups, potentialGroup) // Fügt die 3er-Restgruppe hinzu.
			for _, s := range potentialGroup {
				usedStudents[s] = true
			}
			currentlyUngrouped = []string{}
		}
	}

	var finalUngrouped []string
	for _, student := range allStudents {
		if !usedStudents[student] {
			finalUngrouped = append(finalUngrouped, student)
		}
	}
	return groups, finalUngrouped
}

// ############################################################################################
// checkSymmetricConstraints prüft, ob alle in der Konfiguration definierten
// 		Einschränkungen (Constraints) symmetrisch sind.
// Das bedeutet, wenn Schüler A nicht mit Schüler B zusammenarbeiten soll,
// 		muss auch Schüler B explizit angeben, dass er nicht mit Schüler A zusammenarbeiten soll.
func checkSymmetricConstraints(constraints map[string][]string) error {
	var asymmetricIssues []string // Sammelt alle gefundenen Asymmetrien.

	for studentA, forbiddenByA := range constraints { 
		// Iteriert über jeden Schüler mit Constraints.
		// Sortiere die Liste der verbotenen Schüler für studentA, um konsistente Fehlermeldungen zu gewährleisten.
		sortedForbiddenByA := make([]string, len(forbiddenByA))
		copy(sortedForbiddenByA, forbiddenByA)
		sort.Strings(sortedForbiddenByA)

		for _, studentB := range sortedForbiddenByA { 
			// Iteriert über jeden Schüler, der für studentA verboten ist.
			forbiddenByB, exists := constraints[studentB] // Holt die Constraints für studentB.
			if !exists {  // Wenn studentB keine Constraints hat, ist es asymmetrisch.
				asymmetricIssues = append(asymmetricIssues,
					fmt.Sprintf("Asymmetrie gefunden: '%s' kann nicht mit '%s' arbeiten, aber '%s' hat keine Constraints.",
						studentA, studentB, studentB))
				continue // Gehe zum nächsten studentB.
			}

			foundAInB := false
			for _, s := range forbiddenByB { // Prüft, ob studentA in der Verbotsliste von studentB ist.
				if s == studentA {
					foundAInB = true
					break
				}
			}

			if !foundAInB { // Wenn studentA nicht in der Verbotsliste von studentB gefunden wurde, ist es asymmetrisch.
				asymmetricIssues = append(asymmetricIssues,
					fmt.Sprintf("Asymmetrie gefunden: '%s' kann nicht mit '%s' arbeiten, aber '%s' kann mit '%s' arbeiten.",
						studentA, studentB, studentB, studentA))
			}
		}
	}

	if len(asymmetricIssues) > 0 { // Wenn Asymmetrien gefunden wurden, gib einen Fehler zurück.
		return fmt.Errorf("Inkonsistenzen in den Constraints gefunden:\n%s",
			strings.Join(asymmetricIssues, "\n")) // Fügt alle Meldungen zusammen.
	}
	return nil // Keine Asymmetrien gefunden.
}


// ############################################################################################
// main ist der Haupteinstiegspunkt des Programms.
func main() {

	fmt.Println()
	fmt.Println(strings.Repeat("=", 62))
	fmt.Println("=== Macht zufällige Gruppen für deine Klasse.")
	fmt.Println()

	configPath := "klasse.toml" // Der Standardname der Konfigurationsdatei.
	config, err := readTomlConfig(configPath) // Versucht, die Konfiguration zu lesen oder zu erstellen.

	if err != nil { // Wenn ein Fehler auftritt (z.B. Datei nicht gefunden und neu erstellt).
		if _, ok := err.(*ConfigFileCreatedError); ok { // Prüft, ob es unser spezieller "Datei erstellt"-Fehler ist.
			fmt.Println(err.Error()) // Gibt die Nachricht aus, dass eine neue Datei erstellt wurde.
			// Warte auf Benutzereingabe, bevor das Programm beendet wird, wenn eine Datei erstellt wurde.
			fmt.Println("\n❗️ Drücke Enter, um das Programm zu beenden und die Konfigurationsdatei zu überprüfen.")
			fmt.Scanln() // Wartet auf Enter-Taste.
			os.Exit(0)   // Beendet das Programm sauber.
		} else { // Wenn es ein anderer, schwerwiegender Fehler beim Laden der Konfiguration ist.
			log.Fatalf("❌ Fehler beim Laden der Konfiguration: %v", err) // Gibt den Fehler aus und beendet das Programm abrupt.
		}
	}

	// Zeigt die geladenen Schüler und Constraints an.
	// fmt.Println("Schülerliste:", config.Schuelerliste)
	// fmt.Println("Constraints:", config.Constraints)

	fmt.Println("\n=== Prüfe unverträgliche Paare auf Symmetrie.")
	err = checkSymmetricConstraints(config.Constraints) // Prüft die Symmetrie der Constraints.
	if err != nil {
		fmt.Printf("❗️ Warnung: Unsymmetrische Paare in klasse.toml gefunden: %v\n", err)
		fmt.Println("Die Gruppierung wird fortgesetzt, aber es wird empfohlen, die Konflikte zu korrigieren.")
	} else {
		fmt.Println("✅ Alle Paare sind symmetrisch. Weiter mit der Gruppierung.")
	}

	const attempts = 1000 // Anzahl der Versuche, Gruppen zu bilden (wegen Zufälligkeit).

	// --- Szenario 1: Einteilung in 2er-Gruppen ---
	fmt.Println()
	fmt.Println(strings.Repeat("=", 62))
	fmt.Println("=== Einteilung in 2er-Gruppen (eventuell mit Anpassung).")
	bestGroups2er := [][]string{}       // Speichert die besten gefundenen 2er-Gruppen.
	bestUngrouped2er := []string{}      // Speichert die ungepaarten Schüler für das beste Ergebnis.
	maxGroupedStudents2er := -1         // Verfolgt die maximale Anzahl erfolgreich gruppierter Schüler.

	for i := 0; i < attempts; i++ { // Wiederholt den Gruppierungsprozess mehrmals.
		currentGroups, currentUngrouped := formTwoPersonGroups(config.Schuelerliste, config.Constraints)
		currentGroupedStudents := len(config.Schuelerliste) - len(currentUngrouped) // Anzahl der gruppierten Schüler in diesem Versuch.

		if currentGroupedStudents > maxGroupedStudents2er { // Wenn dieser Versuch besser war.
			maxGroupedStudents2er = currentGroupedStudents
			bestGroups2er = currentGroups
			bestUngrouped2er = currentUngrouped
			if len(currentUngrouped) == 0 { // Wenn alle Schüler gruppiert wurden, ist dies das beste Ergebnis, Abbruch.
				break
			}
		}
	}
	// Ausgabe der Ergebnisse für Szenario 1.
	if len(bestGroups2er) == 0 && len(bestUngrouped2er) > 0 {
		fmt.Println("❌ Es konnten keine gültigen 2er- oder 3er-Gruppen gebildet werden. Alle Schüler sind ungepaart.")
	} else if len(bestGroups2er) == 0 {
		fmt.Println("❌ Es konnten keine 2er- oder 3er-Gruppen gebildet werden.")
	} else {
		for i, group := range bestGroups2er {
			fmt.Printf("Gruppe %d (%d Personen): %v\n", i+1, len(group), group)
		}
	}
	if len(bestUngrouped2er) > 0 {
		fmt.Printf("❗️ Ungruppierte Schüler: %v\n", bestUngrouped2er)
	} else {
		fmt.Println("✅ Alle Schüler wurden erfolgreich in 2er-Gruppen eingeteilt!")
	}

	// --- Szenario 2: Einteilung in 3er-Gruppen ---
	fmt.Println()
	fmt.Println(strings.Repeat("=", 62))
	fmt.Println("=== Einteilung in 3er-Gruppen (eventuell mit Anpassung).")
	bestGroups3er := [][]string{}
	bestUngrouped3er := []string{}
	maxGroupedStudents3er := -1

	for i := 0; i < attempts; i++ {
		currentGroups, currentUngrouped := formThreePersonGroups(config.Schuelerliste, config.Constraints)
		currentGroupedStudents := len(config.Schuelerliste) - len(currentUngrouped)

		if currentGroupedStudents > maxGroupedStudents3er {
			maxGroupedStudents3er = currentGroupedStudents
			bestGroups3er = currentGroups
			bestUngrouped3er = currentUngrouped
			if len(currentUngrouped) == 0 {
				break
			}
		}
	}
	// Ausgabe der Ergebnisse für Szenario 2.
	if len(bestGroups3er) == 0 && len(bestUngrouped3er) > 0 {
		fmt.Println("❌ Es konnten keine gültigen 3er-Gruppen gebildet werden. Alle Schüler sind ungepaart.")
	} else if len(bestGroups3er) == 0 {
		fmt.Println("❌ Es konnten keine 3er-Gruppen gebildet werden.")
	} else {
		for i, group := range bestGroups3er {
			fmt.Printf("Gruppe %d (%d Personen): %v\n", i+1, len(group), group)
		}
	}
	if len(bestUngrouped3er) > 0 {
		fmt.Printf("❗️ Ungruppierte Schüler: %v\n", bestUngrouped3er)
	} else {
		fmt.Println("✅ Alle Schüler wurden erfolgreich in 3er-Gruppen eingeteilt!")
	}

	// --- Szenario 3: Einteilung in 4er-Gruppen ---
	fmt.Println()
	fmt.Println(strings.Repeat("=", 62))
	fmt.Println("=== Einteilung in 4er-Gruppen (eventuell mit Anpassung).")
	bestGroups4er := [][]string{}
	bestUngrouped4er := []string{}
	maxGroupedStudents4er := -1

	for i := 0; i < attempts; i++ {
		currentGroups, currentUngrouped := formFourPersonGroups(config.Schuelerliste, config.Constraints)
		currentGroupedStudents := len(config.Schuelerliste) - len(currentUngrouped)

		if currentGroupedStudents > maxGroupedStudents4er {
			maxGroupedStudents4er = currentGroupedStudents
			bestGroups4er = currentGroups
			bestUngrouped4er = currentUngrouped
			if len(currentUngrouped) == 0 {
				break
			}
		}
	}
	// Ausgabe der Ergebnisse für Szenario 3.
	if len(bestGroups4er) == 0 && len(bestUngrouped4er) > 0 {
		fmt.Println("❌ Es konnten keine gültigen 4er-Gruppen gebildet werden. Alle Schüler sind ungepaart.")
	} else if len(bestGroups4er) == 0 {
		fmt.Println("❌ Es konnten keine 4er-Gruppen gebildet werden.")
	} else {
		for i, group := range bestGroups4er {
			fmt.Printf("Gruppe %d (%d Personen): %v\n", i+1, len(group), group)
		}
	}
	if len(bestUngrouped4er) > 0 {
		fmt.Printf("❗️ Ungruppierte Schüler: %v\n", bestUngrouped4er)
	} else {
		fmt.Println("✅ Alle Schüler wurden erfolgreich in 4er-Gruppen eingeteilt!")
	}
	
	fmt.Println()
	fmt.Println(strings.Repeat("=", 62))
	fmt.Println()

	// Diese Zeilen sind dafür da, das Konsolenfenster auf Windows offen zu halten,
	// 	wenn das Programm per Doppelklick gestartet wird.
	fmt.Println("\nDrücke Enter, um das Programm zu beenden...")
	fmt.Scanln() // Wartet auf die Eingabe der Enter-Taste durch den Benutzer.
}
