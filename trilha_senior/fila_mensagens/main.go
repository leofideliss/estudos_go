package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"sort"
)

func readFiles(path string) [][]string {
	f, _ := os.Open(path)
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = ';'
	records, _ := reader.ReadAll()
	return records
}

func countCustomersMessage(row []string, nomes map[string]int) {
	_, ok := nomes[row[0]]
	if ok {
		nomes[row[0]]++
	} else {
		nomes[row[0]] = 1
	}
}

func getListNames(row []string, nomesUnicos *[]string) {
	unico := slices.Contains(*nomesUnicos, row[0])
	if !unico {
		if row[2] == "true" && row[1] != "" {
			*nomesUnicos = append(*nomesUnicos, row[0])
		}
	}
}

func main() {
	records := readFiles("mensagens.csv")
	nomesUnicos := make([]string, 0)
	nomes := make(map[string]int) // map funcionou por referência sem ter que passar nenhum modificador

	for _, row := range records[1:] {
		countCustomersMessage(row, nomes)
		getListNames(row, &nomesUnicos)
	}

	fmt.Printf("------------------ NOMES ORDENADOS --------------------- \n")
	chaves := make([]string, 0)
	for k := range nomes {
		chaves = append(chaves, k)
	}
	sort.Slice(chaves, func(i, j int) bool {
		return nomes[chaves[i]] < nomes[chaves[j]]
	})

	for _, nome := range chaves {
		fmt.Println("Nome:", nome, "| Quantidade:", nomes[nome])
	}

	fmt.Printf("------------------ NOMES UNICOS COM MENSAGEM ---------------------- \n")
	sort.Strings(nomesUnicos)
	for _, nome := range nomesUnicos {
		fmt.Println("Nome:", nome)
	}

	fmt.Printf("------------------ CUSTOMER QUE MAIS MANDOU MENSAGENS --------------------- \n")

	lastname := chaves[len(chaves)-1]
	fmt.Printf("Cliente: %s - Quantidade: %d \n", lastname, nomes[lastname])

}

// Lista de mensagens
// - nome do remetente
// - conteúdo da mensagem
// - indicador da mensagem

// Quem mandou mais mensagem
// Uma lista ordenada de remetentes unicos que enviaram pelo menos uma mensage unica e não vazia
