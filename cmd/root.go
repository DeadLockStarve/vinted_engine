/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"net/url"
	"github.com/spf13/cobra"
	"github.com/DeadLockStarve/libvinted"
)

var ApiOrderMap map[string]string

func CheckBannedWords(title string,bl []string) bool {
	if len(bl) == 1 && bl[0] == "" {
		return false
	}
	for _, bstring := range bl {
		if strings.Contains(title,bstring) {
			return true
		}
	}
	return false
}

func CheckNationCode(item libvinted.Item,nation_code string) bool {
	var ret bool
	if nation_code == "" {
		ret = true
	} else if strings.ToLower(item.User.Location.Country.Code) == strings.ToLower(nation_code) {
		ret = true
	}
	return ret
}

func GetItems(cmd *cobra.Command, args []string) {
	domain, _ := cmd.Flags().GetString("website")
	price_from, _ := cmd.Flags().GetString("price_from")
	price_to, _ := cmd.Flags().GetString("price_to")
	currency, _ := cmd.Flags().GetString("currency")
	catalog_id, _ := cmd.Flags().GetString("catalog_id")
	order, _ := cmd.Flags().GetString("order")
	user_agent, _ := cmd.Flags().GetString("user_agent")
	download_images,_ := cmd.Flags().GetBool("download_images")
	nation_code, _ := cmd.Flags().GetString("nation_code")
	words_blacklist, _ := cmd.Flags().GetString("words_blacklist")
	wbl := strings.Split(words_blacklist,",")

	vengine := libvinted.GetEngine(domain,user_agent)
	fmt.Println("Retrieving tokens")
	vengine.GetVintedTokens()
	search_text, _ := cmd.Flags().GetString("search_text")
	fmt.Println("Retrieving search results")
	items := vengine.GetItems(libvinted.ItemsArgs{
		SearchText: search_text,
		PriceFrom: libvinted.StrToFloat32(price_from),
		PriceTo: libvinted.StrToFloat32(price_to),
		CatalogIds: catalog_id,
		Currency: currency,
		Order: ApiOrderMap[order],
	})
	fmt.Println("Retrieving items details")
	for _, item := range items {
		if CheckBannedWords(strings.ToLower(item.Title),wbl) {
			continue
		}
		item.RetrieveDetails(vengine)
		if !CheckNationCode(item,nation_code)  {
			continue
		}
		fmt.Println(item.Title,"[" + fmt.Sprint(item.Price.Current) + "][" + fmt.Sprint(item.Price.Total) + "] [" + strconv.Itoa(item.Catalog.Id) + "] [" + item.User.Location.Country.Code + "] [" + item.Url + "]")
		if download_images {
			item.DownloadPhotos(vengine,"searches/" + url.QueryEscape(search_text))
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vinted_engine",
	Short: "Engine for searching automatically vinted items",
	Args: func(cmd *cobra.Command, args []string) error {
		search_text, _ := cmd.Flags().GetString("search_text")
		price_from, _ := cmd.Flags().GetString("price_from")
		price_to, _ := cmd.Flags().GetString("price_to")
		order, _ := cmd.Flags().GetString("order")
		nation_code, _ := cmd.Flags().GetString("nation_code")
		if search_text == "" {
			return fmt.Errorf("Article must be set")
		} else if ! libvinted.IsStrFloat32(price_from) {
			return fmt.Errorf("Minimum price is set but isn't a valid decimal")
		} else if ! libvinted.IsStrFloat32(price_to) {
			return fmt.Errorf("Minimum price is set but isn't a valid decimal")
		} else if (price_from != "0" && price_to != "0" &&
			libvinted.StrToFloat32(price_from) > libvinted.StrToFloat32(price_to)) {
			return fmt.Errorf("Minimum price can't be greater than maximum price")
		} else if nation_code != "" && len(nation_code) != 2 {
			return fmt.Errorf("Nation code lenght must be 2")
		}
		_, ok := ApiOrderMap[order]
		if order != "" && !ok {
			return fmt.Errorf("Accepted arguments for order are [r]elevance, [h]igh_to_low, [l]ow_to_high, [n]ewest_first")
		}
		return nil
	},
	Run: GetItems,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("search_text","s", "", "article name")
	rootCmd.PersistentFlags().StringP("price_from","f", "0", "minimum price")
	rootCmd.PersistentFlags().StringP("price_to","t", "0", "maximum price")
	rootCmd.PersistentFlags().StringP("currency","c", "EUR", "currency")
	rootCmd.PersistentFlags().StringP("catalog_id","", "", "catalog_id")
	rootCmd.PersistentFlags().StringP("words_blacklist","b", "", "words blacklist to exclude from result, separated by \",\"")
	rootCmd.PersistentFlags().StringP("nation_code","n", "", "article nation code (ex. FR for France)")
	rootCmd.PersistentFlags().StringP("order","o", "", "order")
	rootCmd.PersistentFlags().StringP("website","w", "vinted.com", "site domain")
	rootCmd.PersistentFlags().StringP("user_agent","u", "Mozilla/5.0", "user agent")
	rootCmd.PersistentFlags().BoolP("download_images", "d", false, "download images")
	ApiOrderMap = make(map[string]string)
	ApiOrderMap["r"] = "relevance"
	ApiOrderMap["h"] = "price_high_to_low"
	ApiOrderMap["l"] = "price_low_to_high"
	ApiOrderMap["n"] = "newest_first"
}


