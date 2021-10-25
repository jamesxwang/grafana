package conditions

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/stretchr/testify/require"
)

func evaluatorScenario(t *testing.T, json string, reducedValue float64, datapoints ...float64) bool {
	jsonModel, err := simplejson.NewJson([]byte(json))
	require.Nil(t, err)

	evaluator, err := NewAlertEvaluator(jsonModel)
	require.Nil(t, err)

	return evaluator.Eval(null.FloatFrom(reducedValue))
}

func TestEvaluators(t *testing.T) {
	t.Run("greater then", func(t *testing.T) {
		require.True(t, evaluatorScenario(t, `{"type": "gt", "params": [1] }`, 3))
		require.False(t, evaluatorScenario(t, `{"type": "gt", "params": [3] }`, 1))
	})

	t.Run("less then", func(t *testing.T) {
		require.False(t, evaluatorScenario(t, `{"type": "lt", "params": [1] }`, 3))
		require.True(t, evaluatorScenario(t, `{"type": "lt", "params": [3] }`, 1))
	})

	t.Run("within_range", func(t *testing.T) {
		require.True(t, evaluatorScenario(t, `{"type": "within_range", "params": [1, 100] }`, 3))
		require.False(t, evaluatorScenario(t, `{"type": "within_range", "params": [1, 100] }`, 300))
		require.True(t, evaluatorScenario(t, `{"type": "within_range", "params": [100, 1] }`, 3))
		require.False(t, evaluatorScenario(t, `{"type": "within_range", "params": [100, 1] }`, 300))
	})

	t.Run("outside_range", func(t *testing.T) {
		require.True(t, evaluatorScenario(t, `{"type": "outside_range", "params": [1, 100] }`, 1000))
		require.False(t, evaluatorScenario(t, `{"type": "outside_range", "params": [1, 100] }`, 50))
		require.True(t, evaluatorScenario(t, `{"type": "outside_range", "params": [100, 1] }`, 1000))
		require.False(t, evaluatorScenario(t, `{"type": "outside_range", "params": [100, 1] }`, 50))
	})

	t.Run("no_value", func(t *testing.T) {
		t.Run("should be false if series have values", func(t *testing.T) {
			require.False(t, evaluatorScenario(t, `{"type": "no_value", "params": [] }`, 50))
		})

		t.Run("should be true when the series have no value", func(t *testing.T) {
			jsonModel, err := simplejson.NewJson([]byte(`{"type": "no_value", "params": [] }`))
			require.Nil(t, err)

			evaluator, err := NewAlertEvaluator(jsonModel)
			require.Nil(t, err)

			require.True(t, evaluator.Eval(null.FloatFromPtr(nil)))
		})
	})
}
