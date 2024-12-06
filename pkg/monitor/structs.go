package monitor

type awsConfigKey struct{}
type snsClientKey struct{}
type sqsClientKey struct{}

var ctxConfig awsConfigKey
var ctxSNSClient snsClientKey
var ctxSQSClient sqsClientKey
