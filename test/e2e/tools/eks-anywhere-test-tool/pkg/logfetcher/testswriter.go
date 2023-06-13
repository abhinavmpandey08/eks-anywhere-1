package logfetcher

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	awscodebuild "github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/eks-anywhere/pkg/logger"

	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
)

type testsWriter struct {
	filewriter.FileWriter
}

func newTestsWriter(folderPath string) *testsWriter {
	writer := filewriter.NewWriter(folderPath)
	return &testsWriter{FileWriter: writer}
}

func (w *testsWriter) writeCodeBuild(build *awscodebuild.Build) error {
	if _, err := w.Write(constants.BuildDescriptionFile, []byte(build.String()), filewriter.PersistentFile); err != nil {
		return fmt.Errorf("writing build description: %v", err)
	}

	return nil
}

func (w *testsWriter) writeMessages(allMessages, filteredMessages *bytes.Buffer) error {
	if _, err := w.Write(constants.FailedTestsFile, filteredMessages.Bytes(), filewriter.PersistentFile); err != nil {
		return err
	}

	if _, err := w.Write(constants.LogOutputFile, allMessages.Bytes(), filewriter.PersistentFile); err != nil {
		return err
	}

	return nil
}

func (w *testsWriter) writeTest(testName string, logs []*cloudwatchlogs.OutputLogEvent) error {
	buf := new(bytes.Buffer)
	for _, log := range logs {
		buf.WriteString(*log.Message)
	}

	scanner := bufio.NewScanner(buf)
	testBuf := new(bytes.Buffer)

	for scanner.Scan() {
		line := scanner.Text()
		testBuf.WriteString(line + "\n")

		// If the test passed, discard the log by resetting the buffer
		if strings.HasPrefix(line, "--- PASS:") {
			logger.V(2).Info("Test Passed, discarding log", "TestName", line)
			testBuf = new(bytes.Buffer)
		}

		// If the test failed, write the test logs to the test name file and reset  testBuf
		if strings.HasPrefix(line, "--- FAIL:") {
			testName := line[10:strings.LastIndex(line, " ")]
			logger.V(2).Info("Test Failed, writing logs", "TestName", testName)
			if _, err := w.Write(testName, testBuf.Bytes(), filewriter.PersistentFile); err != nil {
				return err
			}
			testBuf = new(bytes.Buffer)
		}
	}

	return nil
}
