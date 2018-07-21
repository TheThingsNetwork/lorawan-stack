import React from 'react'
import PropTypes from 'prop-types'

const Tab = function ({
  isActive,
  children,
  location,
  ariaLabel,
  className,
}) {
  return (
    <li className={className}>
      <a
        href={location}
        aria-label={ariaLabel}
        aria-curent={isActive}
      >
        {children}
      </a>
    </li>
  )
}

Tab.propTypes = {
  isActive: PropTypes.bool.isRequired,
  ariaLabel: PropTypes.string.isRequired,
  location: PropTypes.string,
}

Tab.defaultProps = {
  isActive: false,
  location: '#',
}

export default Tab