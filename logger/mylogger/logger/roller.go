package logger

import (
	"fmt"

	"github.com/wonksing/go-tutorials/logger/mylogger/logger/port"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/types"
	"github.com/wonksing/go-tutorials/logger/mylogger/logger/wrapper"
)

func RollerFactory(rollerType types.RollerType,
	fileName string, maxSize int, maxBackup, maxAge int, compress bool) (port.Roller, error) {

	switch rollerType {
	case types.LumberjackRoller:
		return wrapper.NewLmberjackRoller(fileName, maxSize, maxBackup, maxAge, compress), nil
	default:
		return nil, fmt.Errorf("unknown roller type: %d", rollerType)
	}
}
