package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/nettyrnp/exch-rates/api"
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/api/sys/repository"
	"github.com/nettyrnp/exch-rates/config"
)

func startCmd(flags []cli.Flag) cli.Command {
	return cli.Command{
		Name:  "start",
		Usage: "Starts the Exchange Rates Service API with a given environment file",
		Flags: flags,
		Action: func(c *cli.Context) error {
			env := c.String("env")
			if env == "" {
				return errors.New("you must specify an environment file")
			}

			fmt.Println("-------------------------------------------------")
			fmt.Printf("loading environment configuration from %s\n", env)
			fmt.Println("-------------------------------------------------")
			conf := config.Load(env)
			conf.Print()
			common.InitLogger(conf)
			switch conf.Protocol {
			case "http":
				api.RunHTTP(conf)
				break
			default:
				return fmt.Errorf("unknown protocol: %s", conf.Protocol)
			}
			return nil
		},
	}
}

func migrateCmd(flags []cli.Flag) cli.Command {
	return cli.Command{
		Name:  "migrate",
		Usage: "Applies db migration scripts to db specified in env file",
		Flags: flags,
		Action: func(c *cli.Context) error {
			env := c.String("env")
			if env == "" {
				return errors.New("you must specify an environment file")
			}

			conf := config.Load(env)
			repo := repository.RDBMSRepository{
				Cfg: repository.Config{
					Driver: conf.RepositoryDriver,
					DSN:    conf.RepositoryDSN,
				},
			}

			if initErr := repo.Init(); initErr != nil {
				return initErr
			}

			err := repo.MigrateDown()
			if err != nil {
				return err
			}
			return repo.MigrateUp()
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Exchange Rates Service"
	app.Usage = "Exchange Rates Service CLI tool for managing the API"
	app.Version = "0.0.1-1"

	basicFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "env, e",
			Usage: "Relative path tho the .env file to load",
		},
	}

	app.Commands = []cli.Command{
		startCmd(basicFlags),
		migrateCmd(basicFlags),
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("error running the application: %s\n", err)
	}
}
