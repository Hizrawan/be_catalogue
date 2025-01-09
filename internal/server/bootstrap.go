package server

import (
	"errors"
	"fmt"
	"time"

	"be20250107/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"be20250107/internal/routes"

	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (s *Server) BeforeStart() {
	migrateDatabase(s)
	if err := s.App.Auth.LoadRevocationList(); err != nil {
		panic(err.Error())
	}
}

func (s *Server) AfterStart() {

}

func (s *Server) RegisterRoutes() []RouteRegister {
	return []RouteRegister{
		routes.RegisterAccountRoutes,
		routes.RegisterAuthRoutes,
		routes.RegisterCatalogueRoutes,
	}
}

func migrateDatabase(s *Server) {
	driver, err := mysql.WithInstance(s.App.DB.DB, &mysql.Config{})
	if err != nil {
		panic(err)
	}

	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithInstance("iofs", d, s.App.Config.DBName, driver)
	if err != nil {
		panic(err)
	}

	version, _, _ := m.Version()
	priorVersion := int(version)

	if s.App.Config.Migration.Migrate {
		newVersion := s.App.Config.Migration.Version
		if s.App.Config.Migration.AllowDrop && newVersion == 0 {
			if err := m.Drop(); err != nil {
				panic(err)
			}
			fmt.Printf("%s Reset database success\n", time.Now().Format("2006/01/02 15:04:05"))
			return
		}

		if err := m.Migrate(uint(newVersion)); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("%s Migration attempt failed with err %v\n", time.Now().Format("2006/01/02 15:04:05"), err)
			if s.App.Config.Migration.RollbackOnError {
				if priorVersion == 0 {
					if err := m.Drop(); err != nil {
						panic(err)
					}
					fmt.Println("Database has been reset")
					return
				}
				failedVersion, _, _ := m.Version()
				lastVersion := int(failedVersion) - 1
				if err := m.Force(lastVersion); err != nil {
					fmt.Println("Rollback attempt failed")
					return
				}

				if step := priorVersion - lastVersion; step != 0 {
					if err := m.Steps(step); err != nil {
						panic(err)
					}
				}
				fmt.Printf("Rollback to version %v success\n", priorVersion)
			} else {
				fmt.Println("Rollback on error is disabled")
				return
			}
		} else if err != nil {
			fmt.Printf("%s Migration error: %v\n", time.Now().Format("2006/01/02 15:04:05"), err)
			return
		}

		runningVersion, _, _ := m.Version()
		fmt.Printf("%s Migration version %d\n", time.Now().Format("2006/01/02 15:04:05"), runningVersion)
	}
}
