package quickcrdb

type WorkerOption func(config *workerConfig)

// Sequential will process Qc sequentially, rather than randomly
func Sequential() WorkerOption {
	return func(config *workerConfig) {
		config.sequential = true
	}
}

// Managers sets the number of manager routines. Default is runtime.NumCPU()
func Managers(threads int) WorkerOption {
	return func(config *workerConfig) {
		config.managerRoutines = threads
	}
}

// Workers sets the number of manager routines. Default is runtime.NumCPU()
func Workers(threads int) WorkerOption {
	return func(config *workerConfig) {
		config.workerRoutines = threads
	}
}
