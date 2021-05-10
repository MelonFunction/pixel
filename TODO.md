Refactor
    - system_file, system_render, system_controls don't have clear goals since system_file also renders.
      This should be moved to system_render. 