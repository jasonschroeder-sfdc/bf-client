package client

import (
  "context"
  "fmt"
  "regexp"
  "strings"
  bfpb "github.com/buildfarm/buildfarm/build/buildfarm/v1test"
  "github.com/golang/protobuf/jsonpb"
)

type Queue struct {
  keys []string
}


func createHash(p string, n int) string {
  return fmt.Sprintf("%s:%d", p, n)
}

// func createName(name string, slots []redis.SlotRange) string {
//   h := Hash(name)
//   var n int
//   for n = 0; !slotRangesContainsSlot(slots, Slot(createHash(h, n))); n++ {
//   }
//   re := regexp.MustCompile(`{[^}]*}`)
//   hash := createHash(h, n)
//   if re.MatchString(name) {
//     return re.ReplaceAllString(name, "{" + createHash(h, n) + "}")
//   }
//   return fmt.Sprintf("{%s}%s", hash, name)
// }

func NewQueue(ctx context.Context, name string) *Queue {
  var keys []string
  // result := c.ClusterShards(ctx)
  // if result.Err() != nil {
  //   keys = append(keys, name)
  // } else {
  //   shards := result.Val()
  //   for _, shard := range shards {
  //     keys = append(keys, createName(name, shard.Slots))
  //   }
  // }
  return &Queue {
    keys: keys,
  }
}

func (q *Queue) Length(ctx context.Context) (int64, error) {
  var sum int64 = 0
  // for _, name := range q.keys {
  //   len := rlen(ctx, c, name)
  //   if len.Err() != nil {
  //     return -1, len.Err()
  //   }
  //   sum += len.Val()
  // }
  return sum, nil
}

func ParsePrequeueName(json string) (*Operation, error) {
  ee := &bfpb.ExecuteEntry{}
  err := jsonpb.Unmarshal(strings.NewReader(json), ee)
  if err != nil {
    return nil, err
  }
  return &Operation {
    Name: ee.OperationName,
    Metadata: ee.RequestMetadata,
  }, nil
}

func ParseQueueName(json string) (*Operation, error) {
  qe := &bfpb.QueueEntry{}
  err := jsonpb.Unmarshal(strings.NewReader(json), qe)
  if err != nil {
    return nil, err
  }
  return &Operation {
    Name: qe.ExecuteEntry.OperationName,
    Metadata: qe.ExecuteEntry.RequestMetadata,
  }, nil
}


