package dbclient

import (
	"context"
	"testing"
)

func TestWithCompanyID(t *testing.T) {
	ctx := context.Background()

	// 未设置时不应返回公司 ID
	if v, ok := companyIDFromCtx(ctx); ok || v != nil {
		t.Fatalf("companyIDFromCtx on plain ctx = (%v, %v), want (nil, false)", v, ok)
	}

	// 设置为 0 视为未设置
	if v, ok := companyIDFromCtx(WithCompanyID(ctx, 0)); ok || v != nil {
		t.Fatalf("companyIDFromCtx with 0 = (%v, %v), want (nil, false)", v, ok)
	}

	// 设置合法值
	ctxWithID := WithCompanyID(ctx, 100)
	v, ok := companyIDFromCtx(ctxWithID)
	if !ok || v != uint(100) {
		t.Fatalf("companyIDFromCtx with 100 = (%v, %v), want (100, true)", v, ok)
	}

	// 派生 context 保留值
	derived := context.WithValue(ctxWithID, "other", "x")
	v, ok = companyIDFromCtx(derived)
	if !ok || v != uint(100) {
		t.Fatalf("companyIDFromCtx on derived ctx = (%v, %v), want (100, true)", v, ok)
	}

	// 覆盖为 0 后视为未设置
	if v, ok := companyIDFromCtx(WithCompanyID(ctxWithID, 0)); ok || v != nil {
		t.Fatalf("companyIDFromCtx after overwrite 0 = (%v, %v), want (nil, false)", v, ok)
	}
}
