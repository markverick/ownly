/**
 * Validate email address
 * @param email Email to validate
 * @returns True if email is valid, false otherwise
 */
export function validateEmail(email: string): boolean {
  return !!email
    .toLowerCase()
    .match(
      /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|.(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/,
    );
}

export function convertEmailToName(email: string): string {
  if (!validateEmail(email)) {
    throw new Error('Invalid email address');
  }

  const [localPart, domain] = email.toLowerCase().split('@');
  const sites = [
    {
      site: 'ucla',
      domains: ['g.ucla.edu', 'ucla.edu', 'cs.ucla.edu', 'engineering.ucla.edu'],
      format: ['edu', 'ucla'],
    },
    {
      site: 'arizona',
      domains: ['arizona.edu', 'cs.arizona.edu', 'email.arizona.edu'],
      format: ['edu', 'arizona'],
    },
    { site: 'memphis', domains: ['memphis.edu'], format: ['edu', 'memphis'] },
  ];
  const site = sites.find((candidate) => candidate.domains.includes(domain));
  const format = site?.format ?? ['nonexist'];
  const namePrefix = site ? `/ndnd/${site.site}` : '/ndnd';
  let formatIndex = 0;
  const domainSuffix = domain
    .split('.')
    .reverse()
    .flatMap((label) => {
      if (format[formatIndex] === label) {
        formatIndex++;
        return [];
      }
      return [encodeURIComponent(label)];
    });

  return [namePrefix, ...domainSuffix, encodeURIComponent(localPart)].join('/');
}
