use ort::{
    session::{Session, builder::GraphOptimizationLevel},
    value::Value,
};
use std::sync::{
    Arc, Mutex,
    atomic::{AtomicUsize, Ordering},
};

pub struct Model {
    sessions: Vec<Arc<Mutex<Session>>>,
    counter: AtomicUsize,
}

impl Model {
    pub fn new(model_path: &str, num_instances: usize) -> Result<Self, ort::Error> {
        let sessions = (0..num_instances)
            .map(|_| {
                let session = Session::builder()?
                    .with_optimization_level(GraphOptimizationLevel::Level3)?
                    .commit_from_file(model_path)?;
                Ok(Arc::new(Mutex::new(session)))
            })
            .collect::<Result<Vec<_>, ort::Error>>()?;

        Ok(Self {
            sessions,
            counter: AtomicUsize::new(0),
        })
    }

    fn get_session(&self) -> Arc<Mutex<Session>> {
        let index = self.counter.fetch_add(1, Ordering::SeqCst) % self.sessions.len();
        self.sessions[index].clone()
    }

    pub fn predict(&self, inputs: &[Value]) -> Result<(Vec<i64>, Vec<f32>), ort::Error> {
        let session = self.get_session();
        let mut session = session.lock().unwrap();
        let session_inputs: Vec<_> = inputs.iter().map(|v| v.into()).collect();
        let outputs = session.run(&session_inputs[..])?;

        let (_, labels) = outputs[0].try_extract_tensor::<i64>()?;
        let (_, probs) = outputs[1].try_extract_tensor::<f32>()?;

        Ok((labels.to_vec(), probs.to_vec()))
    }
}
