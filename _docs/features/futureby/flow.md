# FutureBy through the existing SDK client

Task: local SDK prerequisite linked to `t-monitor-ez8.8.7`; the prerequisite Beads acceptance criteria are authoritative. This is a complete additive SDK component. Specifications, money calculations and application activation remain separate accepted tasks.

## Current and expected behavior

The actual SDK base85cc69de initializes a generated InstrumentsServiceClient on its existing authenticated connection. Public GetInstrumentByFIGI exposes a generic Instrument, which lacks future monetary tick and directional margins. Generated FutureBy and Future already exist, but the SDK has no public FutureBy getter.

Add one public GetFutureByFIGI(ctx,figi) returning the typed Future payload through that same InstrumentsServiceClient. With a supplied future FIGI, the request uses generated FIGI identifier type and the exact identifier; the returned FIGI/UID, lot, currency, price step, monetary tick and both directional MoneyValue margins remain the provider's values. For example, step5/tick10 and buy100/sell120 margins must all survive transport unchanged. This getter performs no conversion or financial validation. Task7 owns interpretation, explicit missing-data policies and snapshot provenance.

A disconnected SDK returns an error without RPC. A server failure remains wrapped and recognizable through its gRPC status. Caller cancellation/deadline reaches the RPC; the getter adds no background context or retry. A response with no Instrument produces a nil payload for the consumer to classify. Existing SDK methods and constructors keep their behavior.

## Acceptance scenarios

The prerequisite's AC-S1–AC-S4 cover an actual generated unary transport proof: exact request/authorization/typed payload; server error; caller cancellation; no RPC while disconnected or after Close. Empty FutureResponse must remain an explicit nil payload rather than becoming default monetary fields. Existing consumers compile unchanged because this is one additive concrete method and no interface is widened.

The test transport is in-process grpc bufconn, with generated UnimplementedInstrumentsServiceServer registration and a unary interceptor handling the real FutureBy full method. This crosses request serialization/metadata/generated client and response decoding without a fake client interface or broker connection. No live broker token, sandbox probe or order operation is involved.

## Resource decision

Baseline: one app-owned SDK connection, caller-bounded unary requests, existing RLock and existing per-message transport limit64MiB. New method adds zero connections, goroutines, timers, storage or polling. Per explicit call: one unary request containing one FIGI and one Future response; no list/pagination/history. Current inactive component contributes zero runtime calls until invoked. Response transport ceiling remains64MiB; downstream task7 will reject invalid payloads and apply existing shared rate limits. SDK does not introduce a second rate coordinator or quota-changing retry. Acquisition/current response does not prove exchange-effective or historical financial validity.

Test resources: one in-process server/connection per sequential test case, fixed1MiB bufconn buffer; stop server, close connection/listener, cancel all test contexts. No persistent growth or OS ports. CPU/race/native execution awaits the shared lease. Native gates and independent audit must precede local acceptance. Local acceptance/commit does not authorize remote publication or make a usable released dependency revision.
