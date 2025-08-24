package com.employeemgmt.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;
import org.springframework.jdbc.datasource.DriverManagerDataSource;

import javax.sql.DataSource;

@Configuration
public class DatabaseConfig {

    @Bean
    @Profile("dev")
    @ConfigurationProperties(prefix = "spring.datasource")
    public DataSource devDataSource() {
        return new DriverManagerDataSource();
    }

    @Bean
    @Profile("prod")
    @ConfigurationProperties(prefix = "spring.datasource")
    public DataSource prodDataSource() {
        return new DriverManagerDataSource();
    }

    @Bean
    @Profile("test")
    public DataSource testDataSource() {
        DriverManagerDataSource dataSource = new DriverManagerDataSource();
        dataSource.setDriverClassName("org.h2.Driver");
        dataSource.setUrl("jdbc:h2:mem:testdb");
        dataSource.setUsername("sa");
        dataSource.setPassword("");
        return dataSource;
    }
}
