package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var Uma = []int{20 + 20, 10, -10, -20} // オカを含めたウマ。得点から純粋な打点を計算するために使う。
var Kaeshiten = 30000

const LeastNumMatches = 8 // 「直近n戦」のn

const MaxQualifiedPlayers = 8 // 予選通過人数

// 得点は10倍で扱う。これは整数にするためである。1000打点=1得点なので、得点は0.1単位である。float で持つと計算に誤差が出る。
// 順位は0オリジンで扱う。

type PlayerInfo struct {
	Name    string       // プレイヤー名
	Matches []*MatchInfo // 対戦情報

	TotalDaten       int        // 打点の合計
	MaxDaten         int        // 最高打点
	MaxDatenMatch    *MatchInfo // 最高打点を出した対戦の対戦情報
	TotalScore       int        // 得点の合計
	AverageScore     float64    // 得点の平均
	RecentTotalScore int        // 直近n戦の得点の合計（n戦していない場合は計算されない）
	Place            []int      // 順位の回数
	TotalPlace       int        // 順位の合計
	AveragePlace     float64    // 順位の平均
	RecentPlace      []int      // 直近n戦の順位の回数
	LastBeginTime    string     // 最終対戦開始時刻
}

var Players = map[string]*PlayerInfo{} // プレイヤー名 -> プレイヤー情報

type MatchInfo struct {
	BeginTime string
	EndTime   string
	Players   []*PlayerInfo
	Scores    []int
	Tag       string
	Paipu     string
}

var Matches = []*MatchInfo{}

var tagFlag *string

func main() {
	tagFlag = flag.String("tag", "", "集計対象とするタグ（部分文字列マッチ）")
	flag.Parse()

	r := bufio.NewReader(os.Stdin)
	lineno := 0
	paipuRegexp := regexp.MustCompile(`^\d{6}-[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}$`)

	for {
		lineno++
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, "\n")
		fields := strings.Split(line, ",")
		if len(fields) != 12 {
			fmt.Fprintf(os.Stderr, "parse error at line %d: not enough fields %d\n", lineno, len(fields))
			os.Exit(1)
		}
		if fields[1] == "終了時間" {
			continue
		}
		if !paipuRegexp.MatchString(fields[11]) {
			fmt.Fprintf(os.Stderr, "parse error at line %d: last field is not paipu\n", lineno)
			os.Exit(1)
		}

		if strings.Contains(fields[10], *tagFlag) {
			processMatch(fields)
		}
	}

	// 開始時刻順に対戦情報をソート
	sort.Slice(Matches, func(i, j int) bool {
		return Matches[i].BeginTime < Matches[j].BeginTime
	})

	for _, player := range Players {
		// 開始時刻順に対戦情報をソート
		sort.Slice(player.Matches, func(i, j int) bool {
			return player.Matches[i].BeginTime < player.Matches[j].BeginTime
		})

		// 直近n戦の集計（n戦してない場合は集計しない）
		if len(player.Matches) >= LeastNumMatches {
			recentFirstIndex := len(player.Matches) - LeastNumMatches

			for _, match := range player.Matches[recentFirstIndex:] {
				for i := range match.Players {
					if match.Players[i] == player {
						player.RecentTotalScore += match.Scores[i]
						player.RecentPlace[i]++
					}
				}
			}
		}

		player.AverageScore = float64(player.TotalScore) / float64(len(player.Matches))
		player.AveragePlace = float64(player.TotalPlace) / float64(len(player.Matches))
	}

	fmt.Printf(`
集計対象: 全%d戦

- **太字のプレイヤー** は、そのメトリックで予選通過したプレイヤーです。
- _斜体のプレイヤー_ は、他のメトリックで予選通過したため、そのメトリックでは順位判定から除外されているプレイヤーです。
`, len(Matches))
	qualifiedPlayers := map[*PlayerInfo]bool{}

	var scoreboard []*PlayerInfo

	// 打点王
	scoreboard = make([]*PlayerInfo, 0, len(Players))
	for _, player := range Players {
		scoreboard = append(scoreboard, player)
	}
	sort.Slice(scoreboard, func(i, j int) bool {
		// 最高打点が大きいプレイヤーが上位
		if scoreboard[i].MaxDaten != scoreboard[j].MaxDaten {
			return scoreboard[i].MaxDaten > scoreboard[j].MaxDaten
		}
		// 同じなら最高打点を出した対戦の開始が早いプレイヤーが上位
		return scoreboard[i].MaxDatenMatch.BeginTime < scoreboard[j].MaxDatenMatch.BeginTime
	})
	fmt.Print(`
打点王

| 順位 | プレイヤー名 | 最高打点 | 当該対戦開始時刻 |
| ---: | :--- | ---: | :---: |
`)
	for i, player := range scoreboard {
		playerName := escapeString(player.Name)
		if i == 0 {
			playerName = boldString(playerName)
			qualifiedPlayers[player] = true
		}
		fmt.Printf("| %d | %s | %d | %s |\n",
			i+1, playerName, player.MaxDaten, player.MaxDatenMatch.BeginTime)
	}

	// 直近n戦の平均順位（要n戦）
	scoreboard = make([]*PlayerInfo, 0, len(Players))
	for _, player := range Players {
		if len(player.Matches) >= LeastNumMatches {
			scoreboard = append(scoreboard, player)
		}
	}
	sort.Slice(scoreboard, func(i, j int) bool {
		// 順位の平均が小さい（=順位が上の）プレイヤーが上位
		if scoreboard[i].AveragePlace != scoreboard[j].AveragePlace {
			return scoreboard[i].AveragePlace < scoreboard[j].AveragePlace
		}
		// 同じなら対戦数の多いプレイヤーが上位
		if len(scoreboard[i].Matches) != len(scoreboard[j].Matches) {
			return len(scoreboard[i].Matches) > len(scoreboard[j].Matches)
		}
		// 同じなら最終対局が早いプレイヤーが上位
		return scoreboard[i].LastBeginTime < scoreboard[j].LastBeginTime
	})
	fmt.Printf(`
平均順位（要%d戦）

| 順位 | プレイヤー名 | 平均順位 | 対戦数 |  1位 |  2位 |  3位 |  4位 |
| ---: | :--- | ---: | ---: | ---: | ---: | ---: | ---: |
`, LeastNumMatches)
	for i, player := range scoreboard {
		playerName := escapeString(player.Name)
		if i == 0 {
			playerName = boldString(playerName)
			qualifiedPlayers[player] = true
		}
		fmt.Printf("| %d | %s | %.2f | %d | %d | %d | %d | %d |\n",
			i+1, playerName, player.AveragePlace+1, len(player.Matches),
			player.Place[0], player.Place[1], player.Place[2], player.Place[3])
	}

	// 直近n戦の平均得点（要n戦）
	scoreboard = make([]*PlayerInfo, 0, len(Players))
	for _, player := range Players {
		if len(player.Matches) >= LeastNumMatches {
			scoreboard = append(scoreboard, player)
		}
	}
	sort.Slice(scoreboard, func(i, j int) bool {
		// 直近n戦の得点の合計が大きいプレイヤーが上位（nで割っておく必要はない）
		if scoreboard[i].RecentTotalScore != scoreboard[j].RecentTotalScore {
			return scoreboard[i].RecentTotalScore > scoreboard[j].RecentTotalScore
		}
		// 同じなら最終対局が早いプレイヤーが上位
		return scoreboard[i].LastBeginTime < scoreboard[j].LastBeginTime
	})
	fmt.Printf(`
直近%d戦の平均得点（要%d戦）

| 順位 | プレイヤー名 | 総得点 | 平均得点 |  1位 |  2位 |  3位 |  4位 |
| ---: | :--- | ---: | ---: | ---: | ---: | ---: | ---: |
`, LeastNumMatches, LeastNumMatches)
	for i, player := range scoreboard {
		playerName := escapeString(player.Name)
		if qualifiedPlayers[player] {
			playerName = italicString(playerName)
		} else if len(qualifiedPlayers) < MaxQualifiedPlayers {
			playerName = boldString(playerName)
			qualifiedPlayers[player] = true
		}
		fmt.Printf("| %d | %s | %.2f | %.2f | %d | %d | %d | %d |\n",
			i+1, playerName, float64(player.RecentTotalScore)/10.0, float64(player.RecentTotalScore)/10.0/LeastNumMatches,
			player.RecentPlace[0], player.RecentPlace[1], player.RecentPlace[2], player.RecentPlace[3])
	}

	// 平均得点
	scoreboard = make([]*PlayerInfo, 0, len(Players))
	for _, player := range Players {
		scoreboard = append(scoreboard, player)
	}
	sort.Slice(scoreboard, func(i, j int) bool {
		// 得点の平均が大きいプレイヤーが上位
		if scoreboard[i].AverageScore != scoreboard[j].AverageScore {
			return scoreboard[i].AverageScore > scoreboard[j].AverageScore
		}
		// 同じなら最終対局が早いプレイヤーが上位
		return scoreboard[i].LastBeginTime < scoreboard[j].LastBeginTime
	})
	fmt.Print(`
（参考）平均得点

| 順位 | プレイヤー名 | 総得点 | 平均得点 | 対戦数 |  1位 |  2位 |  3位 |  4位 | 平均順位 |
| ---: | :--- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
`)
	for i, player := range scoreboard {
		playerName := escapeString(player.Name)
		fmt.Printf("| %d | %s | %.2f | %.2f | %d | %d | %d | %d | %d | %.2f |\n",
			i+1, playerName, float64(player.TotalScore)/10.0, player.AverageScore/10.0,
			len(player.Matches),
			player.Place[0], player.Place[1], player.Place[2], player.Place[3],
			player.AveragePlace+1)
	}
}

func processMatch(fields []string) {
	match := &MatchInfo{
		BeginTime: fields[0],
		EndTime:   fields[1],
		Players:   make([]*PlayerInfo, 4),
		Scores:    make([]int, 4),
		Tag:       fields[10],
		Paipu:     fields[11],
	}

	for i := 0; i < 4; i++ {
		playerName := fields[2+i*2]
		score := parseScore(fields[3+i*2])

		if Players[playerName] == nil {
			Players[playerName] = &PlayerInfo{
				Name:        playerName,
				MaxDaten:    -0x80000000,
				Place:       make([]int, 4),
				RecentPlace: make([]int, 4),
			}
		}

		player := Players[playerName]
		player.Matches = append(player.Matches, match)
		daten := scoreToDaten(score, i)
		player.TotalDaten += daten
		if daten > player.MaxDaten {
			player.MaxDaten = daten
			player.MaxDatenMatch = match
		}
		player.TotalScore += score
		player.Place[i]++
		player.TotalPlace += i
		if player.LastBeginTime == "" || match.BeginTime > player.LastBeginTime {
			player.LastBeginTime = match.BeginTime
		}

		match.Players[i] = player
		match.Scores[i] = score
	}

	Matches = append(Matches, match)
}

func parseScore(s string) int {
	// 小数点以下1桁までしか考慮していない。（雀魂の得点の parse においてはそれで問題ない）
	var result int
	fields := strings.SplitN(s, ".", 2)
	integerPart, _ := strconv.ParseInt(fields[0], 10, 32)
	result = int(integerPart) * 10
	if len(fields) > 1 {
		frac := int(fields[1][0] - '0')
		if fields[0][0] != '-' {
			result += frac
		} else {
			result -= frac
		}
	}
	return result
}

func scoreToDaten(score int, place int) int {
	return (score-Uma[place]*10)*100 + Kaeshiten
}
