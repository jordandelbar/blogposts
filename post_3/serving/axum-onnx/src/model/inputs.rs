use crate::types::AdultIncomeInput;
use ndarray::Array;
use ort::value::Value;

pub fn prepare_inputs(data: &[AdultIncomeInput]) -> Result<Vec<Value>, String> {
    use ort::value::Tensor;
    let batch_size = data.len();

    let make_string_array = |v: Vec<String>| {
        Tensor::from_string_array(([batch_size, 1], &v[..]))
            .map_err(|e| format!("Failed to create string tensor: {}", e))
            .map(|t| t.into())
    };

    let make_float_array = |values: Vec<f32>| {
        Array::from_shape_vec((batch_size, 1), values)
            .map_err(|e| format!("Failed to create array: {}", e))
            .and_then(|arr| {
                Value::from_array(arr)
                    .map_err(|e| format!("Failed to create value: {}", e))
                    .map(|v| v.into())
            })
    };

    Ok(vec![
        make_float_array(data.iter().map(|d| d.age).collect())?,
        make_string_array(data.iter().map(|d| d.workclass.clone()).collect())?,
        make_float_array(data.iter().map(|d| d.fnlwgt).collect())?,
        make_string_array(data.iter().map(|d| d.education.clone()).collect())?,
        make_float_array(data.iter().map(|d| d.education_num).collect())?,
        make_string_array(data.iter().map(|d| d.marital_status.clone()).collect())?,
        make_string_array(data.iter().map(|d| d.occupation.clone()).collect())?,
        make_string_array(data.iter().map(|d| d.relationship.clone()).collect())?,
        make_string_array(data.iter().map(|d| d.race.clone()).collect())?,
        make_string_array(data.iter().map(|d| d.sex.clone()).collect())?,
        make_float_array(data.iter().map(|d| d.capital_gain).collect())?,
        make_float_array(data.iter().map(|d| d.capital_loss).collect())?,
        make_float_array(data.iter().map(|d| d.hours_per_week).collect())?,
        make_string_array(data.iter().map(|d| d.native_country.clone()).collect())?,
    ])
}
