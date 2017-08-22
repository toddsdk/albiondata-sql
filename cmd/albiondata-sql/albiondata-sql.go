package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	nats "github.com/nats-io/go-nats"
	"github.com/pcdummy/albiondata-sql/lib"
	adclib "github.com/regner/albiondata-client/lib"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version string
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "albiondata-sql",
	Short: "albiondata-sql is a NATS to SQL Bridge for the Albion Data Project",
	Long: `Reads data from NATS and pushes it to a SQL Database (MSSQL, MySQL, PostgreSQL and SQLite3 are supported), 
creates one table per Market`,
	Run: doCmd,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.albiondata-sql.yaml")
	rootCmd.PersistentFlags().StringP("dbType", "t", "mysql", "Database type must be one of mssql, mysql, postgresql, sqlite3")
	rootCmd.PersistentFlags().StringP("dbURI", "u", "", "Databse URI to connect to, see: http://jinzhu.me/gorm/database.html#connecting-to-a-database")
	rootCmd.PersistentFlags().StringP("natsURL", "n", "nats://public:notsecure@ingest.albion-data.com:4222", "NATS to connect to")
	viper.BindPFlag("dbType", rootCmd.PersistentFlags().Lookup("dbType"))
	viper.BindPFlag("dbURI", rootCmd.PersistentFlags().Lookup("dbURI"))
	viper.BindPFlag("natsURL", rootCmd.PersistentFlags().Lookup("natsURL"))
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc")

		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("albiondata-sql")

		wd, _ := os.Getwd()
		viper.AddConfigPath(wd)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
	}

	viper.SetEnvPrefix("ADS")
	viper.AutomaticEnv()
}

func updateOrCreateOrder(db *gorm.DB, io *adclib.MarketOrder) error {
	location, err := lib.NewLocationFromId(io.LocationID)
	if err != nil {
		return err
	}
	mo := location.Model()

	// fmt.Printf("Importing: %s\n", io.ItemID)

	if err := db.First(&mo, io.ID).Error; err != nil {
		// Not found
		mo = location.Model()
		mo.ID = uint(io.ID)
		mo.ItemID = io.ItemID
		mo.QualityLevel = int8(io.QualityLevel)
		mo.EnchantmentLevel = int8(io.EnchantmentLevel)
		price := strconv.Itoa(io.Price)
		if len(price) > 4 {
			price = price[:len(price)-4]
			i, _ := strconv.Atoi(price)
			mo.Price = i
		} else {
			mo.Price = 0
		}
		mo.Amount = io.Amount
		mo.AuctionType = io.AuctionType
		t, err := time.Parse(time.RFC3339, io.Expires+"+00:00")
		if err != nil {
			return fmt.Errorf("while parsing the time of order id %d, error was: %s", io.ID, err)
		}
		mo.Expires = t

		// fmt.Printf("%s: Creating %s\n", mo.Location.String(), mo.ItemID)
		if err := db.Create(&mo).Error; err != nil {
			return err
		}
	} else {
		// Found, set updatedAt
		// fmt.Printf("%s: Updateing %s\n", mo.Location.String(), mo.ItemID)
		if err := db.Save(&mo).Error; err != nil {
			return err
		}
	}

	return nil
}

func doCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("Connecting to database: %s\n", viper.GetString("dbType"))
	db, err := gorm.Open(viper.GetString("dbType"), viper.GetString("dbURI"))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	defer db.Close()

	if viper.GetString("dbType") == "mysql" {
		for _, l := range lib.Locations() {
			model := l.Model()
			err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model).Error
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
		}
	} else {
		for _, l := range lib.Locations() {
			model := l.Model()
			if err := db.AutoMigrate(&model).Error; err != nil {
				fmt.Printf("%v\n", err)
				return
			}
		}
	}

	nc, _ := nats.Connect(viper.GetString("natsURL"))
	defer nc.Close()

	marketCh := make(chan *nats.Msg, 64)
	marketSub, err := nc.ChanSubscribe(adclib.NatsMarketOrdersDeduped, marketCh)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	defer marketSub.Unsubscribe()

	for {
		select {
		case msg := <-marketCh:
			order := &adclib.MarketOrder{}
			err := json.Unmarshal(msg.Data, order)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				continue
			}

			go func() {
				err = updateOrCreateOrder(db, order)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
				}
			}()
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
