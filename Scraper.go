/**
 * Copyright 2017 Bigpoint GmbH
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"net"
	"fmt"
	"strings"
	"bytes"
	"io"
)

type Scraper struct {
	Socket string
}

func NewScraper(socket string) Scraper {
	scraper := Scraper{socket}
	return scraper
}

/**
Map structure:
  Key: <string> table
  Value: <string> table
*/
func (t *Scraper) GetTables() (map[string]string, error) {
	tableContent, err := t.readSocketToEnd("show table")
	if err != nil {
		return nil, err
	}

	tables := make(map[string]string)

	for _, line := range strings.Split(tableContent, "\n") {
		lineMap := lineToMap(line)
		if lineMap == nil {
			continue
		}
		if _, ok := lineMap["table"]; ok == false {
			continue
		}

		tables[lineMap["table"]] = lineMap["table"]
	}

	return tables, nil
}

/**
Map structure:
  Key: <string> IP
  Value: map
    Key: <string> Key in line (e.g. gpc0)
    Value: <string> Value in line (e.g. 0)
*/
func (t *Scraper) Scrape(table string) (map[string]map[string]string, error) {
	tableContent, err := t.readSocketToEnd(fmt.Sprintf("show table %s", table))
	if err != nil {
		return nil, err
	}

	scraped := make(map[string]map[string]string)

	for _, line := range strings.Split(tableContent, "\n") {
		lineMap := lineToMap(line)
		if lineMap == nil {
			continue
		}
		if _, ok := lineMap["key"]; ok == false {
			continue
		}
		scraped[lineMap["key"]] = lineMap
	}

	return scraped, nil
}

/**
Map structure:
  Key: <string> Key in line (e.g. gpc0)
  Value: <string> Value in line (e.g. 0)
 */
func lineToMap(line string) map[string]string {
	if len(line) < 2 {
		return nil
	}
	// If it is a commented line
	if string(line[0:2]) == "# " {
		// Remove first 2 chars and make it so, that it gets parsed into a map.
		// Required for table detection
		line = string(line[2:])
		line = strings.Replace(line, ":", "=", -1)
		line = strings.Replace(line, ",", "", -1)
		line = strings.Replace(line, "= ", "=", -1)
	}

	// If String is an ID
	if string(line[0:2]) == "0x" {
		// Strip the leading ID
		line = strings.Trim(strings.SplitN(line, ":", 2)[1], " \t")
	}

	lineMap := make(map[string]string)

	for _, kvPair := range strings.Split(line, " ") {
		kv := strings.SplitN(kvPair, "=", 2)
		if len(kv) < 2 {
			continue
		}
		lineMap[kv[0]] = kv[1]
	}

	return lineMap
}

func (t *Scraper) readSocketToEnd(command string) (string, error) {
	connection, err := net.Dial("unix", t.Socket)
	if err != nil {
		return "", err
	}
	// Make sure we close once out of scope
	defer connection.Close()

	// Write the command and copy the output to a buffer
	connection.Write([]byte(fmt.Sprintf("%s\n", command)))
	var buf bytes.Buffer
	io.Copy(&buf, connection)

	// Return the buffer as a string
	return buf.String(), nil
}
