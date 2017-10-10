package beater

import (
  "fmt"
  "time"
  "strings"

  "github.com/elastic/beats/libbeat/beat"
  "github.com/elastic/beats/libbeat/common"
  "github.com/elastic/beats/libbeat/logp"

  "github.com/bradseefeld/jirabeat/config"

  jira "github.com/andygrunwald/go-jira"
)

type Jirabeat struct {
  done   chan struct{}
  config config.Config
  client beat.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
  config := config.DefaultConfig
  if err := cfg.Unpack(&config); err != nil {
    return nil, fmt.Errorf("Error reading config file: %v", err)
  }

  bt := &Jirabeat{
    done:   make(chan struct{}),
    config: config,
  }
  return bt, nil
}

func (bt *Jirabeat) Run(b *beat.Beat) error {
  logp.Info("jirabeat is running! Hit CTRL-C to stop it.")

  var err error
  bt.client, err = b.Publisher.Connect()
  if err != nil {
    return err
  }

  jiraClient, _ := jira.NewClient(nil, bt.config.Url)

  if (bt.config.Authentication.Username != "" && bt.config.Authentication.Password != "") {
    jiraClient.Authentication.SetBasicAuth(bt.config.Authentication.Username, bt.config.Authentication.Password)
  }

  searchOptions := &jira.SearchOptions {
    StartAt: 0,
    MaxResults: 1,
  }

  var openStatuses []string
  for i := 0; i < len(bt.config.OpenStatuses); i++ {
    openStatuses = append(openStatuses, fmt.Sprintf("\"%s\"", bt.config.OpenStatuses[i]))
  }

  openStatusClause := fmt.Sprintf("status in(%s)", strings.Join(openStatuses, ", "))

  ticker := time.NewTicker(bt.config.Period)
  for {
    select {
    case <-bt.done:
      return nil
    case <-ticker.C:
    }

    totalBugs := search(jiraClient, fmt.Sprintf("issuetype = Bug AND project = %s", bt.config.Project), searchOptions)
    openBugs := search(jiraClient, fmt.Sprintf("%s AND issuetype = Bug AND project = %s", openStatusClause, bt.config.Project), searchOptions)

    fields := common.MapStr {
      "type":       "jirabeat", // TODO: I think we can get this from the args
      "project":    bt.config.Project,
      "bugs": common.MapStr {
        "total": totalBugs,
        "open": openBugs,
      },
    }

    byLabel := common.MapStr { }

    for i := 0; i < len(bt.config.Labels); i++ {
      total := search(jiraClient, fmt.Sprintf("labels = %s AND project = %s", bt.config.Labels[i], bt.config.Project), searchOptions)
      byLabel[bt.config.Labels[i]] = common.MapStr {
        "total": total,
        "open": search(jiraClient, fmt.Sprintf("labels = %s AND project = %s AND %s", bt.config.Labels[i], bt.config.Project, openStatusClause), searchOptions),
      }
    }

    fields["labels"] = byLabel

    event := beat.Event {
      Timestamp: time.Now(),
      Fields: fields,
    }
    bt.client.Publish(event)
    logp.Info("Event sent")
  }
}

func (bt *Jirabeat) Stop() {
  bt.client.Close()
  close(bt.done)
}

func search(jiraClient *jira.Client, jql string, searchOptions *jira.SearchOptions) int {
  var _, resp, err = jiraClient.Issue.Search(jql, searchOptions)
  if (err != nil) {
    logp.Err("There was an error when fetching (%s) from Jira: %v", jql, err)
    return -1
  }

  return resp.Total
}
