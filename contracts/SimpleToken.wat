(module
  ;; NUSA Simple Token Contract
  (memory 1)
  (global $total_supply (mut i64) (i64.const 25000000))
  (global $owner i32 (i32.const 0))
  
  (func $get_total_supply (result i64)
    (global.get $total_supply)
  )
  
  (func $transfer (param $from i32) (param $to i32) (param $amount i64)
    ;; Transfer logic here
    (drop)
  )
  
  (export "get_total_supply" (func $get_total_supply))
  (export "transfer" (func $transfer))
)