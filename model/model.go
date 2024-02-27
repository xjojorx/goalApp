package model

import "time"


type Goal struct {
  Id int `db:"id"`
  Name string `db:"name"`
  StartDate time.Time `db:"start_date"`
  TargetDate time.Time `db:"target_date"`
  CurrAmount int `db:"curr_amount"`
  TargetAmount int `db:"target_amount"`
  Pinned bool `db:"pinned"`
}
