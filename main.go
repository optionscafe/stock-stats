package main

import (
  "fmt"
  "flag"
  "time"
  "errors"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "github.com/tidwall/gjson" 
)

const (
	apiBaseUrl = "https://api.tradier.com/v1"
)   

type Quote struct {
  Date string
  Open float64
  High float64
  Low float64
  Close float64
  Volume int       
} 

type Result struct {
  Start string
  End string  
  Open float64
  Close float64
  Difference float64 
}  
      
//
// Main....
//
func main() { 
    
  var results []Result
  current_time := time.Now().Local()    
    
  // Setup flags
  key := flag.String("key", "", "Tradier API key.")
  symbol := flag.String("symbol", "", "Symbol.")
  start := flag.String("start", "1980-01-01", "Start date.")
  end := flag.String("end", current_time.Format("2006-01-02"), "End Date.")
  daysOut := flag.Int("days", 30, "Number of days out.")
  percentAway := flag.Float64("percent_away", 4.5, "Percent away from starting quote.")  
  flag.Parse()
    
  // Setup http client
  client := &http.Client{}    
  
  // Setup api request
  var url = apiBaseUrl + "/markets/history?symbol=" + *symbol + "&start=" + *start + "&end=" + *end + "&interval=daily"
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("Accept", "application/json")
  req.Header.Set("Authorization", fmt.Sprint("Bearer ", *key))   
 
  res, err := client.Do(req)
      
  if err != nil {
    panic(err)  
  }        
  
  // Close Body
  defer res.Body.Close()    
  
  // Make sure the api responded with a 200
  if res.StatusCode != 200 {
    panic(errors.New(fmt.Sprint("API did not return 200, It returned (/markets/history)", res.StatusCode))) 
  }    
     
  // Read the data we got.
  body, err := ioutil.ReadAll(res.Body)
  
  if err != nil {
    panic(err)
  }      
 
  // Get to the array of data.
  vo := gjson.Get(string(body), "history.day")
  
  if ! vo.Exists() {
    panic("No data returned from the Tradier API.")	
  }
  
  var quotes []Quote 
  
  if err := json.Unmarshal([]byte(vo.String()), &quotes); err != nil {    
    panic(err) 
  }  
  
  // Loop through the quotes 
  for key, row := range quotes {

    // Find the date we are going to compare.   
    end, err := FindEndDate(quotes, key, *daysOut)
   
    if err != nil {
      continue
    }
    
    differnce := end.Close - row.Close
    
    results = append(results, Result{ 
                      Start: row.Date,
                      End: end.Date,  
                      Open: row.Close,
                      Close: end.Close,
                      Difference: differnce,      
                    })    
  }
  
  // Get percent up.
  percentUp := PercentUpStat(results, *percentAway)  
  
  // Get percent down.
  percentDown := PercentDownStat(results, *percentAway)
  
  // Print stats
  PrintStats(*symbol, *percentAway, *daysOut, percentUp, percentDown)
}

//
// Print stats.
//
func PrintStats(symbol string, percentAway float64, daysOut int, percentUp float64, percentDown float64) {
  
  fmt.Println("")
  fmt.Println("************************ Stats *****************************")
  fmt.Println("")
  fmt.Println(symbol, "Gains more than", percentAway, "% in any", daysOut, "day period", fmt.Sprintf("%.2f", percentUp) + "% of the time.")
  fmt.Println("")
  fmt.Println(symbol, "Drops more than", percentAway, "% in any", daysOut, "day period", fmt.Sprintf("%.2f", percentDown) + "% of the time.")  
  fmt.Println("")
  fmt.Println("************************************************************")
  fmt.Println("")
  
}

//
// How often does the stock end up up beyond our percent difference.
//
func PercentUpStat(results []Result, percentAway float64) float64 {
  
  var failed int = 0
  
  for _, row := range results {
   
    if PercentChange(row.Open, row.Close) > percentAway {
      
      failed++;
      
    }
    
  }
  
  // Return the number of times percent down did not work.
  return (float64(failed) / float64(len(results))) * 100
  
}

//
// How often does the stock end up down beyond our percent difference.
//
func PercentDownStat(results []Result, percentAway float64) float64 {
  
  var failed int = 0
  
  for _, row := range results {
   
    if PercentChange(row.Open, row.Close) < (percentAway * -1) {
      
      failed++;
      
    }
    
  }
  
  // Return the number of times percent down did not work.
  return (float64(failed) / float64(len(results))) * 100
  
}

//
// Loop through the array and find end date.
//
func FindEndDate(quotes []Quote, index int, daysOut int) (Quote, error) {
  
  // Get the current date based on the date we passed in.
  now, err := time.Parse("2006-01-02", quotes[index].Date)
  
  if err != nil {
    panic(err)
  }  
  
  out := time.Hour * 24 * time.Duration(daysOut)
  outDate := now.Add(out)
  
  // Lopp through until we find the quote that is daysOut out
  for {
    
    // Date out of range of our outlook in dates
    if (index + 1) > len(quotes) {
      return Quote{}, errors.New("Hit end of index")
    }
    
    futNow, err := time.Parse("2006-01-02", quotes[index].Date)
  
    if err != nil {
      panic(err)
    }     
    
    // See if we have passed our outDate or not.
    if futNow.Unix() > outDate.Unix() {
      return quotes[index-1], nil
    }
    
    // Next index
    index++
    
  }
  
  // Should never get here.
  return Quote{}, nil
  
}

//
// Change calculate what is the percentage increase/decrease from [number1] to [number2]
// For example 60 is 200% increase from 20
// It returns result as float64
//
func PercentChange(before float64, after float64) float64 {
  
  diff := float64(after) - float64(before)
  
  realDiff := diff / float64(before)
  
  percentDiff := 100 * realDiff

  return percentDiff
  
}

/* End File */
