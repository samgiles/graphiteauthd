package main

// Split the buffer by '\n' (0x0A) characters, return an byte[][] of
// indicating each metric, and byte[] of the remaining parts of the buffer
func ParseBuffer(buffer []byte) ([][]byte, []byte) {
	metrics := make([][]byte, 8)

	var metricBufferCapacity uint = 0xff
	metricBuffer := make([]byte, metricBufferCapacity)

	var metricSize uint =  0
	var metricBufferUsage uint = 0
	var totalMetrics int = 0

	for _, b := range buffer {
		if b == '\n' {

			metrics[totalMetrics] = metricBuffer[metricBufferUsage - metricSize:metricBufferUsage]
			totalMetrics++

			if totalMetrics > cap(metrics) {
				newMetrics  := make([][]byte, cap(metrics), (cap(metrics) + 1) * 2)
				copy(newMetrics, metrics)
				metrics = newMetrics
			}

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

	return metrics[:totalMetrics], metricBuffer[metricBufferUsage - metricSize:metricBufferUsage]
}
