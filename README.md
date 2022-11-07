## Running the Analyzer

The analyzer can be run on either individual files or on an entire
directory. To run it on an individual file, you should use the 
`--filePath` command line argument. For instance:

```
go run ast-search.go --filePath sample/goroutines/mem-benchmark.go --output results.csv
```

To run the analyzer on a directory containing .go files, you should
use the `--dirPath` command line argument. For instance:

```
go run ast-search.go --dirPath sample --output results.csv
```

Results of the analysis will be stored in the CSV file given using the 
`--output` command line argument.

The file itself contains the following information:


| Column  | Description                                                             |
|---------|-------------------------------------------------------------------------|
| fileName | The absolute path of the analyzed file                                  |	
| waitGroupDecls | The # of `WaitGroup` declarations                                       |	
| condDecls | The # of `Condition` variable declarations                              |
| onceDecls | The # of `Once` declarations                                            |
| mutexDecls | The # of `Mutex` declarations                                           |
| rwMutexDecls | The # of `RWMutex` declarations                                         |
| lockerDecls | The # of `Locker` declarations                                          |
| waitGroupDone | The # of calls to `Done` on a `WaitGroup`                               |
| waitGroupAdd | The # of calls to `Add` on a `WaitGroup`                                |
| waitGroupWait | The # of calls to `Wait` on a `WaitGroup`                               |
| mutexLock | The # of calls to `Lock` on a `Mutex`                                   |
| mutexUnlock | The # of calls to `Unlock` on a `Mutex`                                 |
| rwMutexLock | The # of calls to `Lock` on a `RWMutex`                                 |
| rwMutexUnlock | The # of calls to `Unlock` on a `RWMutex`                               |
| lockerLock | The # of calls to `Lock` on a `Locker`                                  |
| lockerUnlock | The # of calls to `Unlock` on a `Locker`                                |
| condLock | The # of calls to `Lock` on a `Locker` held by a `Condition` variable   |
| condUnlock | The # of calls to `Unlock` on a `Locker` held by a `Condition` variable |
| condWait | The # of calls to `Wait` on a `Condition` variable                      |
| condSignal | The # of calls to `Signal` on a `Condition` variable                    |
| condBroadcast | The # of calls to `Broadcast` on a `Condition` variable                 |
| condNew | The # of calls to `NewCond`                                             |
| onceDo | The # of calls to `Do` on a `Once`                                      |
| unknownDone | The # of uncategorized calls to `Done`                                  |
| unknownAdd | The # of uncategorized calls to `Add`                                   |
| unknownWait | The # of uncategorized calls to `Wait`                                  |
| unknownLock | The # of uncategorized calls to `Lock`                                  |
| unknownUnlock | The # of uncategorized calls to `Unlock`                                |
| unknownSignal | The # of uncategorized calls to `Signal`                                |
| unknownBroadcast | The # of uncategorized calls to `Broadcast`                             |
| unknownDo | The # of uncategorized calls to `Do`                                    |

Note that calls categorized as "unknown" may be completely unrelated to
concurrency. For instance, a function named `Do`, called on a custom
type, would be categorized as "unknownDo", as would a call on a `Once`
value if the analysis cannot determine a `Once` value is the target.