# SDK TradingSchedules — S0 flow

Canon S0-v1; task `t-monitor-ez8.8.6.7.4`, parent `t-monitor-ez8.8.6.7`. High: additive public SDK contract, authentication, caller cancellation and Close/read-lock compatibility. ROOT planning/base/canon/critic authorization `01a11b00-8848-772f-8dde-8882246e8b49`; source implementation, claim and publication are not authorized by this plan.

## Responsibility and current source

One complete backwards-compatible SDK capability: public typed TradingSchedules call through the existing authenticated RealClient transport. Calendar S1 and later app S2/S3 are separate responsibilities; this component certifies neither historical archive, requested-range completeness nor instrument eligibility. No provider data is needed to prove this native transport contract.

Working copy `/private/tmp/tinkoff-go-trading-schedules-spec7`, private branch `codex/trading-schedules-sdk`. Fresh official master and v0.0.11 both `35b2401dc90251acc242e77c222d8acf7824a3e1`, tree `9d7dfd6f5145f1ff188dbf2651e05060567bfe34`; exact working/immediate PR base master at that SHA. Master contains released35b; merge of old FutureBy PR3 has the same tree as accepted2caf. Future moving master requires explicit actual-base/source/compiled/size rebind, not guessed parent. New isolated clone only; old owned root/branch/refs preserved.

SDK/ancestor AGENTS absent; native Makefile has vet/test but no verify/check-pr-size. Existing SDK stdlib/generated grpc/bufconn native test convention and unchanged dependency set apply. Current methodology and independent High planning/audit gates remain applicable; no new module merely to copy t-monitor testing conventions.

## Native boundary

`caller context + *investapi.TradingSchedulesRequest` → `RealClient.GetTradingSchedules` → existing `instrumentsClient.TradingSchedules` → generated native grpc transport → raw `*investapi.TradingSchedulesResponse` or contextual wrapped error. The SDK does not map schedules into strategy/session domain.

Public signature: `GetTradingSchedules(ctx context.Context, request *investapi.TradingSchedulesRequest) (*investapi.TradingSchedulesResponse, error)`. Existing generated types are the sole DTO authority. Optional Exchange nil differs from present empty string; From/To retain native timestamp presence and caller bounds. Tests use nonnil native request objects with absent/present optional fields; no invented nil-context/request acceptance contract or validation policy.

Follow GetFutureByFIGI/GetBonds: `mu.RLock` through the synchronous call and `defer RUnlock`; check connected before RPC; disconnected returns existing `client not connected`. Create auth context with existing `metadata.NewOutgoingContext(ctx,c.metadata)` and invoke existing instrumentsClient once. Preserve caller context cancellation/deadline and original SDK outgoing metadata replacement, not merge arbitrary caller metadata or use private c.ctx. Return the native response without sorting/filtering/conversion; failures return nil and `fmt.Errorf("failed to get trading schedules: %w",err)`. Preserve cause/status. No getter logging, retry, cache, clock/default period, range validation, connection, scheduler or goroutine.

Constructor/connection initialization already owns instrumentsClient; generated InstrumentsServiceClient already exposes TradingSchedules. No new exported aggregate interface or SDK mock must change. All old concrete RealClient methods, constructors/config/Close, generated service/types and old FutureBy tests remain compatible. Future application S2 builds its request and handles its existing rate coordinator; no app module/vendor change here.

## Literal proof of the complete native consumer

New endpoint-specific client test calls the public getter through an in-process generated grpc client and concrete generated server registration with unary interceptor, as accepted FutureBy tests do. The interceptor inspects the actual native TradingSchedules full method/request/metadata and returns the native payload; no handwritten dependency interface implementation. It is a real serialization/transport boundary, not direct invocation of a fake client method. Old FutureBy helper is hardwired to FutureBy and remains untouched.

Explicit cases: request with all optionals absent; Exchange present empty; Exchange set and From/To fixed `time.Date` UTC values. Independently specified expected proto request/response; exact one RPC and only dummy test Bearer auth, preserve deadline. Rich response contains two exchanges, trading/nontrading dates, overnight bounds, auction/evening/clearing timestamps and an unknown interval type; `proto.Equal` validates raw payload preservation. Empty response remains empty. Absence is not interpreted as holiday/history coverage. No scenario-name-dependent setup or production predicate copied into expected results.

Failures: generated server returns grpc status; `%w` preserves errors.Unwrap/status.Code and contextual message. Caller cancellation and deadline reach a started server, terminate it and the public call, and release held RLock. Prove getter retains lock while RPC is active. Disconnected and after actual Close produce zero additional RPC; an initial connected call establishes the counter. Cleanup cancels, stops server, closes client/listener and observes server completion. No live NewReal constructor, real credentials, DNS/provider call or OS listener; buffer1MiB with existing bufconn.

## Scope, resources, compatibility and integration

Exactly four paths: existing `client/real_client.go`; new `client/real_client_trading_schedules_test.go`; this `flow.md`; adjacent `work.md`. Old FutureBy test/helper/docs, constructor/config/proto/generated interfaces, README/Makefile/go.mod/go.sum immutable. No t-monitor file, module adoption or active caller change. Runtime15–35 add+delete forecast, tests180–280, docs40–80, generated runtime0; hard300 is an actual immediate-base gate, not forecast acceptance.

Current dormant addition adds zero automatic provider calls, storage, startup goroutines or scheduling. Future explicit call sends exactly one request over the existing authenticated connection with existing transport limits, caller-owned deadline and no retry; caller/rate/range admission belongs application integration. Native fixture uses1MiB in-memory bufconn, at most two exchanges/three days/two intervals per selected fixture, fixed case count and bounded5s synchronization (short deadline case1s); no fixed-total timing promise. Existing standard cache/toolchain/dependencies reused offline; no installations. Native CPU/race start notice and exact cache permission route are separately coordinated before gates; old9l0 lease/logs grant nothing now.

Rejected alternatives: second connection/private reflection; fabricated calendar authority; scalar SDK parameter conversion that drops optional presence; shared test-helper refactor expanding scope; generated schema refresh; new retries/cache/default date policy. None is needed for the authorized forwarder.

Readiness after planning: bind this exact canon/source/scope and independent High verdict, send full package ROOT before sole ready claim or source implementation. Later current-source native gates, fixed-state fresh independent Medium audit, measured hard300 and operations evidence; any new SDK export/PR/tag/release permission applies only to the concrete audited result. No release version chosen. S0 release availability, application S2 module adoption and final calendar S3 activation remain distinct; parent6.7 stays OPEN.
