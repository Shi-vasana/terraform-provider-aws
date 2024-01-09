// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	configurationDeletedTimeout = 5 * time.Minute
)

func waitClusterCreated(ctx context.Context, conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafka.ClusterStateCreating},
		Target:  []string{kafka.ClusterStateActive},
		Refresh: statusClusterState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafka.Cluster); ok {
		if state, stateInfo := aws.StringValue(output.State), output.StateInfo; state == kafka.ClusterStateFailed && stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateInfo.Code), aws.StringValue(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafka.ClusterStateDeleting},
		Target:  []string{},
		Refresh: statusClusterState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafka.Cluster); ok {
		if state, stateInfo := aws.StringValue(output.State), output.StateInfo; state == kafka.ClusterStateFailed && stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateInfo.Code), aws.StringValue(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterOperationCompleted(ctx context.Context, conn *kafka.Kafka, arn string, timeout time.Duration) (*kafka.ClusterOperationInfo, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{ClusterOperationStatePending, ClusterOperationStateUpdateInProgress},
		Target:  []string{ClusterOperationStateUpdateComplete},
		Refresh: statusClusterOperationState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafka.ClusterOperationInfo); ok {
		if state, errorInfo := aws.StringValue(output.OperationState), output.ErrorInfo; state == ClusterOperationStateUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.ErrorCode), aws.StringValue(errorInfo.ErrorString)))
		}

		return output, err
	}

	return nil, err
}
