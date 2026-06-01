package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"sort"
)

func main() {
	f, _ := os.Open("mensagens.csv")
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'
	records, _ := reader.ReadAll()
	nomes := make(map[string]int)
	nomesUnicos := make([]string, 0)

	var moreMessages string
	maior := 1

	for _, row := range records[1:] {

		_, ok := nomes[row[0]]
		if ok {
			nomes[row[0]]++
		} else {
			nomes[row[0]] = 1
		}

		if nomes[row[0]] > maior {
			maior = nomes[row[0]]
			moreMessages = row[0]
		}

		unico := slices.Contains(nomesUnicos, row[0])
		if !unico {
			if row[2] == "true" && row[1] != "" {
				nomesUnicos = append(nomesUnicos, row[0])
			}
		}

	}

	for nome, quantidade := range nomes {
		fmt.Println("Nome:", nome, "| Quantidade:", quantidade)

	}
	fmt.Println("Mais mensagens:", moreMessages, "| Quantidade:", maior)

	chaves := make([]string, 0)

	for k := range nomes {
		chaves = append(chaves, k)
	}

	sort.Slice(chaves, func(i, j int) bool {
		return nomes[chaves[i]] < nomes[chaves[j]] // ordem crescente de valor
	})
	fmt.Printf("------------------ \n")

	for _, nome := range chaves {
		fmt.Println("Nome:", nome, "| Quantidade:", nomes[nome])
	}

	fmt.Printf("------------------ \n")
	sort.Strings(nomesUnicos)
	for _, nome := range nomesUnicos {
		fmt.Println("Nome:", nome)
	}
}

// Lista de mensagens
// - nome do remetente
// - conteúdo da mensagem
// - indicador da mensagem

// Quem mandou mais mensagem
// Uma lista ordenada de remetentes unicos que enviaram pelo menos uma mensage unica e não vazia
