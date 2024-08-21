package custom_pool

func (nnp *NetworkPool) goPool() {
	run := func() {
		for {

			resource := <-nnp.loader

			if exit, ok := resource.res.(string); ok && exit == "EXIT" {
				return
			}

			nnp.ActionBox.Check(resource.res)

			if nnp.FuncSignalDone != nil {
				nnp.FuncSignalDone(1)
			}

			nnp.wg.Done()
		}
	}

	for _ = range nnp.Threads {
		go run()
	}
}

func (nnp *NetworkPool) Close() {
	for _ = range nnp.Threads {
		nnp.loader <- loader_resource{res: "EXIT"}
	}

	close(nnp.loader)
}
