package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	rootCmd.PersistentFlags().Int64P("expireCheckEvery", "e", 3600, "every x seconds the db entries get checked if an order is expired" )
	rootCmd.PersistentFlags().Int64P("expireNotUpdFor", "s", 86400, "expires oder if it was not updated for x seconds. Default 1 Day" )

	viper.BindPFlag("dbType", rootCmd.PersistentFlags().Lookup("dbType"))
	viper.BindPFlag("dbURI", rootCmd.PersistentFlags().Lookup("dbURI"))
	viper.BindPFlag("natsURL", rootCmd.PersistentFlags().Lookup("natsURL"))
	viper.BindPFlag("expireCheckEvery", rootCmd.PersistentFlags().Lookup("expireCheckEvery"))
	viper.BindPFlag("expireNotUpdFor", rootCmd.PersistentFlags().Lookup("expireNotUpdFor"))
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

		// Add the executable path as
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		viper.AddConfigPath(exPath)
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
	mo := lib.NewModelMarketOrder()

	// fmt.Printf("Importing: %s\n", io.ItemID)

	if err := db.Unscoped().Where("albion_id = ?", io.ID).First(&mo).Error; err != nil {
		// Not found
		mo = lib.NewModelMarketOrder()
		mo.Location = location
		mo.AlbionID = uint(io.ID)
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
		mo.InitialAmount = io.Amount
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
		mo.Amount = io.Amount
		mo.DeletedAt = nil
		if err := db.Save(&mo).Error; err != nil {
			return err
		}
	}

	return nil
}

func expireOrders(db *gorm.DB) {
	checkEvery := viper.GetInt64("expireCheckEvery")
	expNotUpdFor := -viper.GetInt64("expireNotUpdFor")

	for {
		now := time.Now()
		nowE := now.Add(time.Duration(expNotUpdFor)* time.Second)
		if err := db.Table(lib.NewModelMarketOrder().TableName()).Where("expires <= ? or updated_at <= ?", now, nowE).Update(map[string]interface{}{"deleted_at": now}).Error; err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}

		time.Sleep(time.Second * time.Duration(checkEvery))
	}
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
		model := lib.NewModelMarketOrder()
		err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model).Error
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
	} else {
		model := lib.NewModelMarketOrder()
		if err := db.AutoMigrate(&model).Error; err != nil {
			fmt.Printf("%v\n", err)
			return
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

	if viper.GetInt64("expireCheckEvery") > 0 {
		go expireOrders(db)
	}

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
