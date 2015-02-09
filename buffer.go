package main

// Split the buffer by '\n' (0x0A) characters, return an byte[][] of
// indicating each metric, and byte[] of the remaining parts of the buffer
func ParseBuffer(buffer []byte) ([][]byte, []byte) {
	metrics := make([][]byte, 0)

	var metricBufferCapacity uint32 = 0xff
	metricBuffer := make([]byte, metricBufferCapacity)

	var metricSize uint32 =  0
	var metricBufferUsage uint32 = 0

	for _, b := range buffer {
		if b == '\n' {

			metrics = append(metrics, metricBuffer[metricBufferUsage - metricSize:metricBufferUsage])
			metricSize = 0;
		} else {

			if metricBufferUsage == metricBufferCapacity {
				newMetricBufferCapacity := (metricBufferCapacity + 1) * 2
				newBuffer := make([]byte, metricBufferCapacity, newMetricBufferCapacity)
				copy(newBuffer, metricBuffer)
				metricBuffer = newBuffer
				metricBufferCapacity = newMetricBufferCapacity
			}

			metricBuffer[metricBufferUsage] = b
			metricSize++
			metricBufferUsage++
		}
	}

	return metrics, metricBuffer[:metricSize]
}
