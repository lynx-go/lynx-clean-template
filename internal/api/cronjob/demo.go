package cronjob

import (
	"context"

	"github.com/lynx-go/lynx/contrib/schedule"
	"github.com/lynx-go/x/log"
)

type DemoTask struct {
}

func (t *DemoTask) Name() string {
	return "DemoTask"
}

func (t *DemoTask) Cron() string {
	return "@every 5s"
}

func (t *DemoTask) HandlerFunc() schedule.HandlerFunc {
	return func(ctx context.Context) error {
		log.InfoContext(ctx, "hello from demo task")
		return nil
	}
}

var _ schedule.Task = new(DemoTask)

func NewDemoTask() *DemoTask {
	return &DemoTask{}
}
