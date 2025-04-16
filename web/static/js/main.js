document.addEventListener("DOMContentLoaded", () => {
  console.log("Digital Marketplace loaded");

  const observerOptions = {
    root: null, // relative to document viewport
    rootMargin: '0px',
    threshold: 0.1 // trigger when 10% of the element is visible
  };

  const observerCallback = (entries, observer) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target); // Stop observing once visible
      }
    });
  };

  const observer = new IntersectionObserver(observerCallback, observerOptions);

  const elementsToAnimate = document.querySelectorAll('.slide-in-up');
  elementsToAnimate.forEach(el => {
    observer.observe(el);
  });
});
