// Mock for until-async to avoid ES module parsing issues
// This is a minimal mock that provides the 'until' function that MSW uses

module.exports = {
  until: function until(condition, options = {}) {
    const {
      timeout = 5000,
      interval = 50,
    } = options;

    return new Promise((resolve, reject) => {
      const startTime = Date.now();
      
      const check = () => {
        try {
          if (condition()) {
            resolve();
            return;
          }
          
          if (Date.now() - startTime >= timeout) {
            reject(new Error('until condition timeout'));
            return;
          }
          
          setTimeout(check, interval);
        } catch (error) {
          reject(error);
        }
      };
      
      check();
    });
  },
};


