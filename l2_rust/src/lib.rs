use wasmtime::*;

pub struct NusaVM {
    engine: Engine,
    store: Store<()>,
}

impl NusaVM {
    pub fn new() -> Self {
        println!("Initializing NUSA VM...");
        let engine = Engine::default();
        let store = Store::new(&engine, ());
        NusaVM { engine, store }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_vm_creation() {
        let vm = NusaVM::new();
        assert!(true);
    }
}
