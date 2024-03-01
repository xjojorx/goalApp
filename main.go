package main

import (
	"goalApp/model"
	"goalApp/templ"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func routes(e *echo.Echo) {
  e.GET("/", func(c echo.Context) error {
    goals, err := AllGoals(db)
    if err != nil {
      return err
    }
    component := templ.Home(goals)
    return component.Render(c.Request().Context(), c.Response().Writer)
  })

  e.GET("/goal", func(c echo.Context) error {
    qparam := c.QueryParams()
    idStr := qparam.Get("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
      return c.String(http.StatusBadRequest, "Invalid id")
    }

    goal, err := GetGoalById(db, id)
    if err != nil {
      c.NoContent(404)
    }

    comp := templ.Goal(goal, true)
    return comp.Render(c.Request().Context(), c.Response().Writer)
  })
  e.PATCH("/goal", func(c echo.Context) error {
    qparam := c.QueryParams()
    idStr := qparam.Get("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
      log.Printf("error getting id, received '%s'", idStr)
      return c.String(http.StatusBadRequest, "Invalid id")
    }

    qttyStr := c.FormValue("changeAmount")
    qtty, err := strconv.Atoi(qttyStr)
    if err != nil {
      log.Printf("error getting amount, received '%s'", qttyStr)
      log.Printf("%+v", c.Request())
      return c.String(http.StatusBadRequest, "Invalid amount")
    }

    op := qparam.Get("op")

    goal, err := GetGoalById(db, id)
    if err != nil {
      return c.NoContent(404)
    }

    if op == "reduce" {
      goal.CurrAmount -= qtty
    }
    if op == "set" {
      goal.CurrAmount = qtty
    }
    if op == "add" {
      goal.CurrAmount += qtty
    }

    updated, err := UpdateGoal(db, goal)
    if err != nil {
      return err
    }

    comp := templ.Goal(updated, true)
    return comp.Render(c.Request().Context(), c.Response().Writer)
  })

  e.GET("/goals", func(c echo.Context) error {
    goals, err := AllGoals(db)
    if err != nil {
      return c.String(http.StatusInternalServerError, "Error retrieving goals")
    }

    comp := templ.Goals(goals)
    return comp.Render(c.Request().Context(), c.Response().Writer)
  })

  e.GET("/modal-goal", func(c echo.Context) error {
    comp := templ.FormModalGoal()
    return comp.Render(c.Request().Context(), c.Response().Writer)
  })

  e.POST("/goal", func(c echo.Context) error {
    name := c.FormValue("Name")
    CurrAmount := c.FormValue("CurrAmount")
    TargetAmount := c.FormValue("TargetAmount")
    StartDate := c.FormValue("StartDate")
    TargetDate := c.FormValue("TargetDate")

    curr, err := strconv.Atoi(CurrAmount)
    if err != nil {
      log.Printf("error parsing current amount, received: %s, %#v", CurrAmount, err)
      curr = 0
    }
    target, err := strconv.Atoi(TargetAmount)
    if err != nil {
      log.Printf("error parsing target amount, received: %s, %#v", TargetAmount, err)
      return err
    }
    start, err := time.Parse("2006-01-02", StartDate)
    if err != nil {
      log.Printf("error parsing start date, received: %s, %#v", StartDate, err)
      start = time.Now().UTC()
    }
    end, err := time.Parse("2006-01-02",TargetDate)
    if err != nil {
      log.Printf("error parsing target date, received: %s, %#v", TargetDate, err)
      return err
    }

    goal := model.Goal{
      Name: name,
      CurrAmount: curr,
      TargetAmount: target,
      StartDate: start,
      TargetDate: end,
    }

    goal, err = CreateGoal(db, goal)
    if err != nil {
      return err
    }

    //c.Response().Header().Add("HX-Trigger", "refreshBar")
    // Add() and Set() mess with header case
    // c.Response().Header()["HX-Trigger"] =[]string{ "refreshBar"}

    comp := templ.Goal(goal, true)
    return comp.Render(c.Request().Context(), c.Response().Writer)
  })
}

var db *sqlx.DB


func AllGoals(db *sqlx.DB) ([]model.Goal, error) {
  goals := []model.Goal{}
  err := db.Select(&goals, "SELECT * FROM goals")
  if err != nil {
    log.Println("error: ", err)
    return nil, err
  }
  
  return goals, nil
}

func GetGoalById(db *sqlx.DB, goalId int) (goal model.Goal, e error) {
  e = db.Get(&goal, "SELECT * FROM goals WHERE id = ?", goalId)
  if e != nil {
    log.Printf("Error getting goal with id %d, error: %s\n", goalId, e)
  }

  return
}

func CreateGoal(db *sqlx.DB, goal model.Goal) (resGoal model.Goal, e error) {
  query := "INSERT INTO goals (name, start_date, target_date, curr_amount, target_amount, pinned) VALUES (:name, :start_date, :target_date, :curr_amount, :target_amount, :pinned)"

  res, err := db.NamedExec(query, goal)
  if err != nil {
    log.Println("error: ", err)
    e = err
    return
  }
  insertedId, err := res.LastInsertId()
  log.Printf("inserted on id '%d'", insertedId)
  
  resGoal, err = GetGoalById(db, int(insertedId))

  return
}

func UpdateGoal(db *sqlx.DB, goal model.Goal) (resGoal model.Goal, e error) {
  query := "UPDATE goals set name=:name, start_date=:start_date, target_date=:target_date, curr_amount=:curr_amount, target_amount=:target_amount, pinned=:pinned WHERE id=:id"

  _, err := db.NamedExec(query, goal)
  if err != nil {
    log.Println("error: ", err)
    e = err
    return
  }
  log.Printf("updated goal on id '%d'", goal.Id)
  
  resGoal, err = GetGoalById(db, goal.Id)

  return
}

func main() {
  var err error
  db, err = sqlx.Connect("sqlite3", "data.db")
  if err != nil {
    log.Fatalln(err)
  }


  // goal := model.Goal{
  //   Name: "go goal utc",
  //   StartDate: time.Now().UTC(),
  //   TargetDate: time.Now().UTC().AddDate(1, 0, 0),
  //   TargetAmount: 1234,
  // }
  // g, er := CreateGoal(db, goal)
  // if er != nil {
  //   log.Fatal(er)
  // }
  // log.Printf("Created %#v\n", g)

  e := echo.New()
  e.Use(middleware.Logger())

  routes(e)

  e.Static("/images", "images")
  e.Static("/css", "css")
  e.Logger.Fatal(e.Start(":42069"))
}
