package main

// Split the buffer by '\n' (0x0A) characters, return an byte[][] of
// indicating each metric, and byte[] of the remaining parts of the buffer
func ParseBuffer(buffer []byte) ([][]byte, []byte) {
	metrics := make([][]byte, 0)
	metricBuffer := make([]byte, 0xff)

	var metricSize uint32 =  0

	for _, b := range buffer {
		if b == '\n' {

			newMetric := make([]byte, metricSize)
			copy(newMetric, metricBuffer[:metricSize])

			metrics = append(metrics, newMetric)
			metricSize = 0;
		} else {
			metricBuffer[metricSize] = b;
			metricSize++;
		}
	}

	return metrics, metricBuffer[:metricSize]
}
