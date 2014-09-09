Reverse-GoL
===========

This is my Kaggle entry for the "Reverse Game-of-Life" competition that ended in May 2014.  

The general approach is to use Genetic Algorithms to generate precursor board(s) that target the desired end-state.  
As a sub-goal, I used this as an opportunity to create something self-contained in GoLang.

In general, however, the top of the leaderboard was dominated by people with the computing power (and smarts) to 
enumerate the precursor boards exhaustively.  

Therefore, while my approach is interesting, it's main value is in the 'lessons learned' : And, of course, 
whatever useful code you can find in this project.

I also have to emphasise that this project was coded rather rapidly, particularly in the end-stages - so there 
are definitely a lot of rough edges (although some parts are quite refined - it's just really uneven...)


Presentation
-------------------

There's a presentation about the GoLang / Kaggle experience at http://RedCatLabs.com/2014-06-05_Reverse-GoL/


Installation
-------------------

```
git clone <ThisRepo>
cd <ThisRepo>
GOPATH=`pwd` go build reverse-gol.go speed_packed.go ga.go board-standard.go transitions.go db.go && ./reverse-gol
```

Installation of MySQL library : 

```
GOPATH=`pwd` go get github.com/go-sql-driver/mysql
```

Then, (for definiteness, obviously you can change these defaults by editing ```db.go:48```) create a database user 'reverse-gol' with password 'reverse-gol' with access rights to database 'reverse-gol'.

The MySQL table creation commands are at the beginning of ```db.go``` - and you'll have to run these manually yourself.



Running
-------------------
To compile and run, use the following :

```
GOPATH=`pwd` go build reverse-gol.go speed_packed.go ga.go board-standard.go transitions.go db.go && ./reverse-gol
```

To see the different use-cases of this only-built-for-results code, do a ```./reverse-gol --help```, and then examine the source...

```
Usage:
  -cmd="": Required : {db|create|visualize|run|submit}
  -count=0: Number of ids to process
  -delta=0: Number of steps between start and end
  -id=0: Specific id to examine
  -seed=1: Random seed to use
  -training=false: Act on training set (default=false, i.e. test set)
  -type="": create:{fake_training_data|training_set_transitions|synthetic_transitions}, db:{test|insert_problems}, visualize:{data|ga}, submit:{kaggle|fakescore}
```
