package com.employeemgmt.service;

import com.employeemgmt.model.Employee;
import com.employeemgmt.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

@Service
@Transactional
public class EmployeeService {

    @Autowired
    private EmployeeRepository employeeRepository;

    public List<Employee> getAllEmployees() {
        return employeeRepository.findAll();
    }

    public Page<Employee> getAllEmployees(Pageable pageable) {
        return employeeRepository.findAll(pageable);
    }

    public Optional<Employee> getEmployeeById(Long id) {
        return employeeRepository.findById(id);
    }

    public Optional<Employee> getEmployeeByEmail(String email) {
        return employeeRepository.findByEmail(email);
    }

    public Employee saveEmployee(Employee employee) {
        // Check if email already exists for new employees
        if (employee.getId() == null && employeeRepository.findByEmail(employee.getEmail()).isPresent()) {
            throw new IllegalArgumentException("Employee with email " + employee.getEmail() + " already exists");
        }
        return employeeRepository.save(employee);
    }

    public Employee updateEmployee(Long id, Employee employeeDetails) {
        Optional<Employee> existingEmployee = employeeRepository.findById(id);
        if (existingEmployee.isEmpty()) {
            throw new IllegalArgumentException("Employee not found with id: " + id);
        }

        Employee employee = existingEmployee.get();

        // Check if email is being changed and if new email already exists
        if (!employee.getEmail().equals(employeeDetails.getEmail())) {
            Optional<Employee> emailExists = employeeRepository.findByEmail(employeeDetails.getEmail());
            if (emailExists.isPresent() && !emailExists.get().getId().equals(id)) {
                throw new IllegalArgumentException("Employee with email " + employeeDetails.getEmail() + " already exists");
            }
        }

        employee.setFirstName(employeeDetails.getFirstName());
        employee.setLastName(employeeDetails.getLastName());
        employee.setEmail(employeeDetails.getEmail());
        employee.setPhone(employeeDetails.getPhone());
        employee.setDepartment(employeeDetails.getDepartment());
        employee.setPosition(employeeDetails.getPosition());
        employee.setHireDate(employeeDetails.getHireDate());
        employee.setSalary(employeeDetails.getSalary());

        return employeeRepository.save(employee);
    }

    public void deleteEmployee(Long id) {
        if (!employeeRepository.existsById(id)) {
            throw new IllegalArgumentException("Employee not found with id: " + id);
        }
        employeeRepository.deleteById(id);
    }

    public List<Employee> getEmployeesByDepartment(String department) {
        return employeeRepository.findByDepartment(department);
    }

    public List<Employee> getEmployeesByPosition(String position) {
        return employeeRepository.findByPosition(position);
    }

    public List<Employee> searchEmployeesByName(String name) {
        return employeeRepository.findByNameContaining(name);
    }

    public List<Employee> getEmployeesByDepartmentAndPosition(String department, String position) {
        return employeeRepository.findByDepartmentAndPosition(department, position);
    }

    public List<String> getAllDepartments() {
        return employeeRepository.findAllDepartments();
    }

    public List<String> getAllPositions() {
        return employeeRepository.findAllPositions();
    }

    public Long getEmployeeCountByDepartment(String department) {
        return employeeRepository.countByDepartment(department);
    }
}
